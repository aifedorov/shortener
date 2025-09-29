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
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewSaveJSONHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		userID         string
		urlCheckerErr  error
		storeErr       error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful URL shortening",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/abc123"}`,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"url": "https://example.com"`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "empty request body",
			requestBody:    ``,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "invalid URL",
			requestBody:    `{"url": "invalid-url"}`,
			userID:         "user123",
			urlCheckerErr:  errors.New("invalid URL"),
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "conflict error",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"result":"http://localhost:8080/existing"}`,
		},
		{
			name:           "repository error",
			requestBody:    `{"url": "https://example.com"}`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty URL in JSON",
			requestBody:    `{"url": ""}`,
			userID:         "user123",
			urlCheckerErr:  errors.New("empty URL"),
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "missing URL field",
			requestBody:    `{}`,
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/abc123"}`,
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

			if tt.userID != "" {
				if tt.requestBody != "" && !strings.Contains(tt.requestBody, "invalid") && !strings.Contains(tt.requestBody, `"url": "invalid-url"`) {
					var reqBody RequestBody
					if err := json.Unmarshal([]byte(tt.requestBody), &reqBody); err == nil {
						mockURLChecker.EXPECT().CheckURL(reqBody.URL).Return(tt.urlCheckerErr)
						if tt.urlCheckerErr == nil {
							mockRepo.EXPECT().Store(tt.userID, cfg.BaseURL, reqBody.URL).Return("http://localhost:8080/abc123", tt.storeErr)
						}
					}
				} else if strings.Contains(tt.requestBody, `"url": "invalid-url"`) {
					mockURLChecker.EXPECT().CheckURL("invalid-url").Return(tt.urlCheckerErr)
				} else if strings.Contains(tt.requestBody, `"url": ""`) {
					mockURLChecker.EXPECT().CheckURL("").Return(tt.urlCheckerErr)
				} else if tt.requestBody == `{}` {
					mockURLChecker.EXPECT().CheckURL("").Return(nil)
					mockRepo.EXPECT().Store(tt.userID, cfg.BaseURL, "").Return("http://localhost:8080/abc123", tt.storeErr)
				}
			}

			handler := NewSaveJSONHandler(cfg, mockRepo, mockURLChecker)

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
