package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewSaveJSONBatchHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		userID         string
		urlCheckerErr  error
		storeBatchErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful batch URL shortening",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}, {"correlation_id": "2", "original_url": "https://google.com"}]`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[{"correlation_id":"1","short_url":"http://localhost:8080/abc1"},{"correlation_id":"2","short_url":"http://localhost:8080/abc2"}]`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty request body",
			requestBody:    ``,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "invalid URL in batch",
			requestBody:    `[{"correlation_id": "1", "original_url": "invalid-url"}]`,
			userID:         "user123",
			urlCheckerErr:  errors.New("invalid URL"),
			storeBatchErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "conflict error",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"result":"http://localhost:8080/existing"}`,
		},
		{
			name:           "repository error",
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty batch array",
			requestBody:    `[]`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[]`,
		},
		{
			name:           "missing correlation_id",
			requestBody:    `[{"original_url": "https://example.com"}]`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeBatchErr:  nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `[{"correlation_id":"","short_url":"http://localhost:8080/abc1"}]`,
		},
		{
			name:           "missing original_url",
			requestBody:    `[{"correlation_id": "1"}]`,
			userID:         "user123",
			urlCheckerErr:  errors.New("empty URL"),
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
			mockURLChecker := mocks.NewMockURLChecker(ctrl)

			if tt.userID != "" && tt.requestBody != "" && !strings.Contains(tt.requestBody, "invalid") {
				var batchReqs []BatchRequest
				if err := json.Unmarshal([]byte(tt.requestBody), &batchReqs); err == nil {
					for _, req := range batchReqs {
						if req.OriginalURL != "" {
							mockURLChecker.EXPECT().CheckURL(req.OriginalURL).Return(tt.urlCheckerErr)
						} else if tt.name == "missing original_url" {
							mockURLChecker.EXPECT().CheckURL("").Return(tt.urlCheckerErr)
						}
					}
					if tt.urlCheckerErr == nil {
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
			} else if tt.userID != "" && strings.Contains(tt.requestBody, "invalid-url") {
				mockURLChecker.EXPECT().CheckURL("invalid-url").Return(tt.urlCheckerErr)
			} else if tt.name == "unauthorized user" {
				mockURLChecker.EXPECT().CheckURL("https://example.com").Return(nil)
			}

			handler := NewSaveJSONBatchHandler(cfg, mockRepo, mockURLChecker)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
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
