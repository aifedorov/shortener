package handlers

import (
	"context"
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

func TestNewSavePlainTextHandler(t *testing.T) {
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
			requestBody:    "https://example.com",
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc123",
		},
		{
			name:           "invalid URL",
			requestBody:    "invalid-url",
			userID:         "user123",
			urlCheckerErr:  errors.New("invalid URL"),
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    "https://example.com",
			userID:         "",
			urlCheckerErr:  nil,
			storeErr:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "conflict error",
			requestBody:    "https://example.com",
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   "http://localhost:8080/existing",
		},
		{
			name:           "repository error",
			requestBody:    "https://example.com",
			userID:         "user123",
			urlCheckerErr:  nil,
			storeErr:       errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty request body",
			requestBody:    "",
			userID:         "user123",
			urlCheckerErr:  errors.New("empty URL"),
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
			mockURLChecker := mocks.NewMockURLChecker(ctrl)

			if tt.userID != "" {
				mockURLChecker.EXPECT().CheckURL(tt.requestBody).Return(tt.urlCheckerErr)
				if tt.urlCheckerErr == nil {
					mockRepo.EXPECT().Store(tt.userID, cfg.BaseURL, tt.requestBody).Return("http://localhost:8080/abc123", tt.storeErr)
				}
			}

			handler := NewSavePlainTextHandler(cfg, mockRepo, mockURLChecker)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "text/plain")

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestNewSavePlainTextHandler_ReadBodyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	mockRepo := mocks.NewMockRepository(ctrl)
	mockURLChecker := mocks.NewMockURLChecker(ctrl)

	handler := NewSavePlainTextHandler(cfg, mockRepo, mockURLChecker)

	req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
	req.Header.Set("Content-Type", "text/plain")

	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Bad Request\n", rr.Body.String())
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
