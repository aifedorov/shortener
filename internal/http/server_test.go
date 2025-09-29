package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
)

func TestServer_Integration(t *testing.T) {
	t.Parallel()

	t.Run("full workflow", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)

		// Setup expectations
		mockRepo.EXPECT().Store(gomock.Any(), gomock.Any(), "https://example.com").Return("http://localhost:8080/abc123", nil).Times(2)
		mockRepo.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return([]repository.URLOutput{
			{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
		}, nil)
		mockRepo.EXPECT().Ping().Return(nil)

		server := NewServer(config.NewConfig(), mockRepo)
		server.mountHandlers()

		userID := uuid.NewString()
		ctx := context.WithValue(context.Background(), auth.UserIDKey, userID)

		// 1. Create short URL via plain text
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com")).WithContext(ctx)
		req.Header.Set("Content-Type", "text/plain")
		res := executeRequest(req, server)
		assert.Equal(t, http.StatusCreated, res.Code)

		shortURL := res.Body.String()
		assert.NotEmpty(t, shortURL)

		// 2. Create short URL via JSON
		req = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "https://example.com"}`)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		res = executeRequest(req, server)
		assert.Equal(t, http.StatusCreated, res.Code)

		// 3. Get user URLs
		req = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil).WithContext(ctx)
		res = executeRequest(req, server)
		assert.Equal(t, http.StatusOK, res.Code)

		// 4. Test ping
		req = httptest.NewRequest(http.MethodGet, "/ping", nil)
		res = executeRequest(req, server)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("server initialization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cfg := config.NewConfig()
		mockRepo := mocks.NewMockRepository(ctrl)
		server := NewServer(cfg, mockRepo)

		assert.NotNil(t, server)
		assert.NotNil(t, server.router)
		assert.Equal(t, cfg, server.config)
		assert.Equal(t, mockRepo, server.repo)
	})

	t.Run("middleware integration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		server := NewServer(config.NewConfig(), mockRepo)
		server.mountHandlers()

		// Test without user ID (should return unauthorized for plain text)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		res := executeRequest(req, server)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func TestServer_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("invalid routes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().Get("nonexistent").Return("", repository.ErrShortURLNotFound)
		server := NewServer(config.NewConfig(), mockRepo)
		server.mountHandlers()

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		res := executeRequest(req, server)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		server := NewServer(config.NewConfig(), mockRepo)
		server.mountHandlers()

		req := httptest.NewRequest(http.MethodPut, "/", nil)
		res := executeRequest(req, server)
		assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	})
}

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)
	return r
}
