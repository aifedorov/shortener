package handlers

import (
	"context"
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

func TestNewSavePlainTextHandler(t *testing.T) {
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
			requestBody:    "https://example.com",
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc123",
		},
		{
			name:           "invalid URL",
			requestBody:    "invalid-url",
			userID:         "user123",
			storeErr:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request\n",
		},
		{
			name:           "unauthorized user",
			requestBody:    "https://example.com",
			userID:         "",
			storeErr:       nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "conflict error",
			requestBody:    "https://example.com",
			userID:         "user123",
			storeErr:       repository.NewConflictError("http://localhost:8080/existing", repository.ErrURLExists),
			expectedStatus: http.StatusConflict,
			expectedBody:   "http://localhost:8080/existing",
		},
		{
			name:           "repository error",
			requestBody:    "https://example.com",
			userID:         "user123",
			storeErr:       errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty request body",
			requestBody:    "",
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

			// Create real domain service (lightweight, no external dependencies)
			validator := validate.NewService()
			randomizer := random.NewService()
			urlService := urlDomain.NewService(randomizer, validator)

			if tt.userID != "" && tt.requestBody != "" && !strings.Contains(tt.requestBody, "invalid") {
				mockRepo.EXPECT().Store(tt.userID, cfg.BaseURL, tt.requestBody).Return("http://localhost:8080/abc123", tt.storeErr)
			}

			handler := NewSavePlainTextHandler(cfg, mockRepo, urlService)

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

	// Create real domain service (lightweight, no external dependencies)
	validator := validate.NewService()
	randomizer := random.NewService()
	urlService := urlDomain.NewService(randomizer, validator)

	handler := NewSavePlainTextHandler(cfg, mockRepo, urlService)

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

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}
