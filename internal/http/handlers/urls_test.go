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

func TestNewURLsHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		getAllResult   []repository.URLOutput
		getAllError    error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful get URLs",
			userID: "user123",
			getAllResult: []repository.URLOutput{
				{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
				{ShortURL: "http://localhost:8080/def456", OriginalURL: "https://google.com"},
			},
			getAllError:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com"},{"short_url":"http://localhost:8080/def456","original_url":"https://google.com"}]`,
		},
		{
			name:           "user has no URLs",
			userID:         "user123",
			getAllResult:   nil,
			getAllError:    repository.ErrUserHasNoData,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "No Content\n",
		},
		{
			name:           "unauthorized user",
			userID:         "",
			getAllResult:   nil,
			getAllError:    nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "repository error",
			userID:         "user123",
			getAllResult:   nil,
			getAllError:    errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name:           "empty URLs list",
			userID:         "user123",
			getAllResult:   []repository.URLOutput{},
			getAllError:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name:   "single URL",
			userID: "user123",
			getAllResult: []repository.URLOutput{
				{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
			},
			getAllError:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com"}]`,
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

			if tt.userID != "" {
				mockRepo.EXPECT().GetAll(tt.userID, cfg.BaseURL).Return(tt.getAllResult, tt.getAllError)
			}

			handler := NewURLsHandler(cfg, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if strings.HasPrefix(tt.expectedBody, "[") {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			} else {
				assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
			}
		})
	}
}
