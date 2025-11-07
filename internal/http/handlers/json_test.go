package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/config"
	urlDomain "github.com/aifedorov/shortener/internal/domain/url"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewSaveJSONHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		userID         string
		storeErr       error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful URL shortening",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/abc123"}`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"url": "https://example.com"`,
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty request body",
			requestBody:    ``,
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "invalid URL",
			requestBody:    `{"url": "invalid-url"}`,
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "",
			storeErr:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "conflict error",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			storeErr:       repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"result":"http://localhost:8080/existing"}`,
		},
		{
			name:           "repository error",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			storeErr:       errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty URL in JSON",
			requestBody:    `{"url": ""}`,
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "missing URL field",
			requestBody:    `{}`,
			userID:         "user123",
			storeErr:       nil,
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

			if tt.userID != "" {
				if tt.requestBody != "" && !strings.Contains(tt.requestBody, "invalid") && !strings.Contains(tt.requestBody, `"url": "invalid-url"`) && tt.requestBody != `{}` {
					var reqBody RequestBody
					if err := json.Unmarshal([]byte(tt.requestBody), &reqBody); err == nil && reqBody.URL != "" {
						mockRepo.EXPECT().Store(tt.userID, cfg.BaseURL, reqBody.URL).Return("http://localhost:8080/abc123", tt.storeErr)
					}
				}
			}

			handler := NewSaveJSONHandler(cfg, mockRepo, urlService)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if strings.HasPrefix(tt.expectedBody, "{") {
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
