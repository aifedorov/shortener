package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/config"
	urlDomain "github.com/aifedorov/shortener/internal/domain/url"
	userDomain "github.com/aifedorov/shortener/internal/domain/user"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewSaveJSONBatchHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		userID         string
		storeBatchErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful batch URL shortening",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}, {"correlation_id": "2", "original_url": "https://google.com"}]`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[{"correlation_id":"1","short_url":"http://localhost:8080/abc1"},{"correlation_id":"2","short_url":"http://localhost:8080/abc2"}]`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty request body",
			requestBody:    ``,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "invalid URL in batch",
			requestBody:    `[{"correlation_id": "1", "original_url": "invalid-url"}]`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "",
			storeBatchErr:  nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "conflict error",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "user123",
			storeBatchErr:  repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"result":"http://localhost:8080/existing"}`,
		},
		{
			name:           "repository error",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "user123",
			storeBatchErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty batch array",
			requestBody:    `[]`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[]`,
		},
		{
			name:           "missing correlation_id",
			requestBody:    `[{"original_url": "https://example.com"}]`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[{"correlation_id":"","short_url":"http://localhost:8080/abc1"}]`,
		},
		{
			name:           "missing original_url",
			requestBody:    `[{"correlation_id": "1"}]`,
			userID:         "user123",
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := &config.Config{
				BaseURL: "http://localhost:8080",
			}

			mockRepo := mocks.NewMockRepository(ctrl)
			validator := validate.NewService()
			randomizer := random.NewService()
			urlService := urlDomain.NewService(randomizer, validator)
			userService := userDomain.NewService()

			if tt.userID != "" && tt.requestBody != "" && !strings.Contains(tt.requestBody, "invalid") {
				var batchReqs []BatchRequest
				if err := json.Unmarshal([]byte(tt.requestBody), &batchReqs); err == nil {
					// Check if all URLs are non-empty
					allValid := true
					for _, req := range batchReqs {
						if req.OriginalURL == "" {
							allValid = false
							break
						}
					}
					if allValid {
						urls := make([]repository.BatchURLInput, len(batchReqs))
						for i, req := range batchReqs {
							urls[i] = repository.BatchURLInput{
								CID:         req.CID,
								OriginalURL: req.OriginalURL,
							}
						}
						results := make([]repository.BatchURLOutput, len(urls))
						for i, url := range urls {
							results[i] = repository.BatchURLOutput{
								CID:      url.CID,
								ShortURL: fmt.Sprintf("http://localhost:8080/abc%d", i+1),
							}
						}
						mockRepo.EXPECT().StoreBatch(tt.userID, cfg.BaseURL, urls).Return(results, tt.storeBatchErr)
					}
				}
			}

			handler := NewSaveJSONBatchHandler(cfg, mockRepo, urlService, userService)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := userService.SetUserIDToContext(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if strings.HasPrefix(tt.expectedBody, "[") || strings.HasPrefix(tt.expectedBody, "{") {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
			}
		})
	}
}
