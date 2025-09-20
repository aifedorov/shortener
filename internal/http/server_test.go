package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/aifedorov/shortener/internal/pkg/random"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/auth"
	"github.com/aifedorov/shortener/internal/repository"
)

func TestServer_redirect(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		expectedContentType string
		expectedCode        int
		expectedLocation    string
		expectedBody        string
	}
	tests := []struct {
		name   string
		server *Server
		method string
		path   string
		want   want
	}{
		{
			name:   "Get method without id",
			server: server,
			method: http.MethodGet,
			path:   `/`,
			want: want{
				expectedContentType: "text/plain; charset=utf-8",
				expectedCode:        http.StatusBadRequest,
				expectedBody:        fmt.Sprintf("%s\n", ErrShortURLMissing.Error()),
			},
		},
		{
			name: "Get method with existing id",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			),
			method: http.MethodGet,
			path:   `/1`,
			want: want{
				expectedContentType: "text/html; charset=utf-8",
				expectedCode:        http.StatusTemporaryRedirect,
				expectedLocation:    "https://google.com",
			},
		},
		{
			name: "Get method with not existing id",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			),
			method: http.MethodGet,
			path:   `/2`,
			want: want{
				expectedContentType: "text/plain; charset=utf-8",
				expectedCode:        http.StatusNotFound,
				expectedBody:        "404 page not found\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.expectedCode, res.Code)
			assert.Equal(t, tt.want.expectedContentType, res.Header().Get("Content-Type"))

			if tt.want.expectedLocation != "" {
				assert.Equal(t, tt.want.expectedLocation, res.Header().Get("Location"))
			}

			if tt.want.expectedBody != "" {
				assert.Equal(t, tt.want.expectedBody, res.Body.String())
			}
		})
	}
}

func TestServer_saveURL_TextPlain(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name:        "Post method without parameters",
			server:      server,
			method:      http.MethodPost,
			contentType: "text/plain",
			want: want{
				contentType: "text/plain",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with valid url",
			server:      server,
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			want: want{
				contentType: "text/plain",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with invalid url",
			server:      server,
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `bad_data`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name: "Post method with existed url",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					Rand: random.NewService(),
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			),
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			want: want{
				contentType: "text/plain",
				code:        http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, res.Body.String())
			}
		})
	}
}

func TestServer_saveURL_JSON(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name:        "Post with empty JSON",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{}`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with valid JSON",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://practicum.yandex.ru"}`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with invalid JSON",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://practicum.yandex.ru}`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name:        "Post method with invalid URL parameter",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "bad_data"}`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name: "Post method with existed URL",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					Rand: random.NewService(),
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			),
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://google.com"}`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, res.Body.String())
			}
		})
	}
}

func TestServer_deleteURLs(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name:        "Delete method with valid JSON",
			server:      server,
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `["abc123", "def456"]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusAccepted,
			},
		},
		{
			name:        "Delete method with invalid JSON",
			server:      server,
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `["abc123", "def456"`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name:        "Delete method with empty aliases",
			server:      server,
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `[]`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name: "Delete method with deleted URL",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					PathToURL: map[string]string{
						"abc123": "https://google.com",
					},
				},
			),
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `["abc123"]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusAccepted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/user/urls", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
		})
	}
}

func TestNewPingHandler(t *testing.T) {
	t.Parallel()

	type want struct {
		code int
	}
	tests := []struct {
		name string
		cfg  *config.Config
		want want
	}{
		{
			name: "ping in memory storage",
			cfg:  &config.Config{},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "ping file storage",
			cfg: &config.Config{
				FileStoragePath: "tmp",
			},
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.cfg, repository.NewRepository(context.Background(), tt.cfg))
			server.mountHandlers()

			req := httptest.NewRequest(http.MethodGet, "/ping", strings.NewReader(""))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			res := executeRequest(req, server)

			assert.Equal(t, tt.want.code, res.Code)
		})
	}
}

func TestServer_getURLs(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name   string
		server *Server
		method string
		want   want
	}{
		{
			name:   "Get method with no URLs in repository",
			server: server,
			method: http.MethodGet,
			want: want{
				contentType: "application/json",
				code:        http.StatusOK,
				body:        `[]`,
			},
		},
		{
			name: "Get method with URLs in repository",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				&repository.MemoryRepository{
					PathToURL: map[string]string{
						"abc123": "https://google.com",
						"def456": "https://yandex.ru",
					},
				},
			),
			method: http.MethodGet,
			want: want{
				contentType: "application/json",
				code:        http.StatusOK,
				body:        `[{"short_url":"http://localhost:8080/abc123","original_url":"https://google.com"},{"short_url":"http://localhost:8080/def456","original_url":"https://yandex.ru"}]`,
			},
		},
		{
			name: "Get method with single URL in repository",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				&repository.MemoryRepository{
					PathToURL: map[string]string{
						"abc123": "https://google.com",
					},
				},
			),
			method: http.MethodGet,
			want: want{
				contentType: "application/json",
				code:        http.StatusOK,
				body:        `[{"short_url":"http://localhost:8080/abc123","original_url":"https://google.com"}]`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/user/urls", nil)
			ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
			req = req.WithContext(ctx)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			if tt.want.body != "" {
				if strings.HasPrefix(tt.want.body, "[") {
					var expected []map[string]interface{}
					var actual []map[string]interface{}

					err := json.Unmarshal([]byte(tt.want.body), &expected)
					if err != nil {
						t.Fatalf("Failed to unmarshal expected JSON: %v", err)
					}

					err = json.Unmarshal(res.Body.Bytes(), &actual)
					if err != nil {
						t.Fatalf("Failed to unmarshal actual JSON: %v", err)
					}

					assert.Equal(t, len(expected), len(actual))

					for _, expectedItem := range expected {
						found := false
						for _, actualItem := range actual {
							if assert.ObjectsAreEqual(expectedItem, actualItem) {
								found = true
								break
							}
						}
						assert.True(t, found, "Expected item not found in actual response: %v", expectedItem)
					}
				} else {
					assert.JSONEq(t, tt.want.body, res.Body.String())
				}
			}
		})
	}
}

type MockRepository struct {
	PathToURL   map[string]string
	Rand        *random.Service
	mu          sync.RWMutex
	conflictURL string
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		PathToURL: make(map[string]string),
		Rand:      random.NewService(),
	}
}

func (m *MockRepository) Run() error   { return nil }
func (m *MockRepository) Ping() error  { return nil }
func (m *MockRepository) Close() error { return nil }

func (m *MockRepository) Get(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if url, exists := m.PathToURL[shortURL]; exists {
		return url, nil
	}
	return "", repository.ErrShortURLNotFound
}

func (m *MockRepository) GetAll(userID, baseURL string) ([]repository.URLOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]repository.URLOutput, len(m.PathToURL))
	i := 0
	for alias, url := range m.PathToURL {
		res[i] = repository.URLOutput{
			ShortURL:    baseURL + "/" + alias,
			OriginalURL: url,
		}
		i++
	}
	return res, nil
}

func (m *MockRepository) Store(userID, baseURL, targetURL string) (string, error) {
	if m.conflictURL != "" && targetURL == m.conflictURL {
		alias := "existing123"
		shortURL := baseURL + "/" + alias
		return "", repository.NewConflictError(shortURL, repository.ErrURLExists)
	}

	alias, err := m.Rand.GenRandomString()
	if err != nil {
		return "", err
	}

	shortURL := baseURL + "/" + alias
	m.mu.Lock()
	m.PathToURL[alias] = targetURL
	m.mu.Unlock()

	return shortURL, nil
}

func (m *MockRepository) StoreBatch(userID, baseURL string, urls []repository.BatchURLInput) ([]repository.BatchURLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	res := make([]repository.BatchURLOutput, len(urls))
	for i, url := range urls {
		if m.conflictURL != "" && url.OriginalURL == m.conflictURL {
			alias := "existing123"
			shortURL := baseURL + "/" + alias
			return nil, repository.NewConflictError(shortURL, repository.ErrURLExists)
		}

		alias, err := m.Rand.GenRandomString()
		if err != nil {
			return nil, err
		}

		shortURL := baseURL + "/" + alias
		res[i] = repository.BatchURLOutput{
			CID:      url.CID,
			ShortURL: shortURL,
		}
		m.PathToURL[alias] = url.OriginalURL
	}

	return res, nil
}

func (m *MockRepository) DeleteBatch(userID string, aliases []string) error {
	if len(aliases) == 0 {
		return errors.New("aliases is empty")
	}

	m.mu.Lock()
	for _, alias := range aliases {
		delete(m.PathToURL, alias)
	}
	m.mu.Unlock()
	return nil
}

func TestServer_saveURL_TextPlain_Conflict(t *testing.T) {
	t.Parallel()

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name: "Post method with conflicting URL",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				func() *MockRepository {
					mock := NewMockRepository()
					mock.conflictURL = "https://google.com"
					return mock
				}(),
			),
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			want: want{
				contentType: "text/plain",
				code:        http.StatusConflict,
				body:        "http://localhost:8080/existing123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveURL_JSON_Conflict(t *testing.T) {
	t.Parallel()

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name: "Post method with conflicting URL",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				func() *MockRepository {
					mock := NewMockRepository()
					mock.conflictURL = "https://google.com"
					return mock
				}(),
			),
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://google.com"}`,
			want: want{
				contentType: "application/json",
				code:        http.StatusConflict,
				body:        `{"result":"http://localhost:8080/existing123"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveBatchURLs_Conflict(t *testing.T) {
	t.Parallel()

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name: "Post method with conflicting URL in batch",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				func() *MockRepository {
					mock := NewMockRepository()
					mock.conflictURL = "https://google.com"
					return mock
				}(),
			),
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"}]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusConflict,
				body:        `{"result":"http://localhost:8080/existing123"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten/batch", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveURL_TextPlain_Unauthorized(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		noUserID    bool
		want        want
	}{
		{
			name:        "Post method without user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			noUserID:    true,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
		{
			name:        "Post method with empty user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			noUserID:    false,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))

			if !tt.noUserID {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, "")
				req = req.WithContext(ctx)
			}

			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveURL_JSON_Unauthorized(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		noUserID    bool
		want        want
	}{
		{
			name:        "Post method without user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://google.com"}`,
			noUserID:    true,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
		{
			name:        "Post method with empty user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://google.com"}`,
			noUserID:    false,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.requestBody))

			if !tt.noUserID {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, "")
				req = req.WithContext(ctx)
			}

			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_getURLs_Unauthorized(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name     string
		server   *Server
		method   string
		noUserID bool
		want     want
	}{
		{
			name:     "Get method without user ID in context",
			server:   server,
			method:   http.MethodGet,
			noUserID: true,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
		{
			name:     "Get method with empty user ID in context",
			server:   server,
			method:   http.MethodGet,
			noUserID: false,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusUnauthorized,
				body:        "Unauthorized\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/user/urls", nil)

			if !tt.noUserID {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, "")
				req = req.WithContext(ctx)
			}

			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveBatchURLs_Unauthorized(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		noUserID    bool
		want        want
	}{
		{
			name:        "Post method without user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"}]`,
			noUserID:    true,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusInternalServerError,
				body:        "Internal Server Error\n",
			},
		},
		{
			name:        "Post method with empty user ID in context",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"}]`,
			noUserID:    false,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusInternalServerError,
				body:        "Internal Server Error\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten/batch", strings.NewReader(tt.requestBody))

			if !tt.noUserID {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, "")
				req = req.WithContext(ctx)
			}

			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_deleteURLs_Unauthorized(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		noUserID    bool
		want        want
	}{
		{
			name:        "Delete method without user ID in context",
			server:      server,
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `["abc123"]`,
			noUserID:    true,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusInternalServerError,
				body:        "Internal Server Error\n",
			},
		},
		{
			name:        "Delete method with empty user ID in context",
			server:      server,
			method:      http.MethodDelete,
			contentType: "application/json",
			requestBody: `["abc123"]`,
			noUserID:    false,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusInternalServerError,
				body:        "Internal Server Error\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/user/urls", strings.NewReader(tt.requestBody))

			if !tt.noUserID {
				ctx := context.WithValue(req.Context(), auth.UserIDKey, "")
				req = req.WithContext(ctx)
			}

			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

type MockRepositoryForNoContent struct {
	PathToURL    map[string]string
	Rand         *random.Service
	mu           sync.RWMutex
	returnNoData bool
}

func NewMockRepositoryForNoContent() *MockRepositoryForNoContent {
	return &MockRepositoryForNoContent{
		PathToURL: make(map[string]string),
		Rand:      random.NewService(),
	}
}

func (m *MockRepositoryForNoContent) Run() error   { return nil }
func (m *MockRepositoryForNoContent) Ping() error  { return nil }
func (m *MockRepositoryForNoContent) Close() error { return nil }

func (m *MockRepositoryForNoContent) Get(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if url, exists := m.PathToURL[shortURL]; exists {
		return url, nil
	}
	return "", repository.ErrShortURLNotFound
}

func (m *MockRepositoryForNoContent) GetAll(userID, baseURL string) ([]repository.URLOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.returnNoData {
		return nil, repository.ErrUserHasNoData
	}

	res := make([]repository.URLOutput, len(m.PathToURL))
	i := 0
	for alias, url := range m.PathToURL {
		res[i] = repository.URLOutput{
			ShortURL:    baseURL + "/" + alias,
			OriginalURL: url,
		}
		i++
	}
	return res, nil
}

func (m *MockRepositoryForNoContent) Store(userID, baseURL, targetURL string) (string, error) {
	alias, err := m.Rand.GenRandomString()
	if err != nil {
		return "", err
	}

	shortURL := baseURL + "/" + alias
	m.mu.Lock()
	m.PathToURL[alias] = targetURL
	m.mu.Unlock()

	return shortURL, nil
}

func (m *MockRepositoryForNoContent) StoreBatch(userID, baseURL string, urls []repository.BatchURLInput) ([]repository.BatchURLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	res := make([]repository.BatchURLOutput, len(urls))
	for i, url := range urls {
		alias, err := m.Rand.GenRandomString()
		if err != nil {
			return nil, err
		}

		shortURL := baseURL + "/" + alias
		res[i] = repository.BatchURLOutput{
			CID:      url.CID,
			ShortURL: shortURL,
		}
		m.PathToURL[alias] = url.OriginalURL
	}

	return res, nil
}

func (m *MockRepositoryForNoContent) DeleteBatch(userID string, aliases []string) error {
	if len(aliases) == 0 {
		return errors.New("aliases is empty")
	}

	m.mu.Lock()
	for _, alias := range aliases {
		delete(m.PathToURL, alias)
	}
	m.mu.Unlock()
	return nil
}

func TestServer_getURLs_NoContent(t *testing.T) {
	t.Parallel()

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name   string
		server *Server
		method string
		want   want
	}{
		{
			name: "Get method when user has no URLs",
			server: NewServer(
				&config.Config{BaseURL: "http://localhost:8080"},
				func() *MockRepositoryForNoContent {
					mock := NewMockRepositoryForNoContent()
					mock.returnNoData = true
					return mock
				}(),
			),
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusNoContent,
				body:        "No Content\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/user/urls", nil)
			ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
			req = req.WithContext(ctx)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, res.Body.String())
		})
	}
}

func TestServer_saveBatchURLs(t *testing.T) {
	t.Parallel()

	server := NewServer(config.NewConfig(), repository.NewMemoryRepository())

	type want struct {
		contentType string
		code        int
		body        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		contentType string
		requestBody string
		want        want
	}{
		{
			name:        "Post method with valid batch JSON",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"}, {"correlation_id": "2", "original_url": "https://yandex.ru"}]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with empty batch array",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with invalid JSON",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name:        "Post method with invalid URL",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "invalid_url"}]`,
			want: want{
				contentType: "text/plain; charset=utf-8",
				code:        http.StatusBadRequest,
			},
		},
		{
			name:        "Post method with missing correlation_id",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"original_url": "https://google.com"}]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with missing original_url",
			server:      server,
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1"}]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
		{
			name: "Post method with existing URLs",
			server: NewServer(
				config.NewConfig(),
				&repository.MemoryRepository{
					Rand: random.NewService(),
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			),
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `[{"correlation_id": "1", "original_url": "https://google.com"}]`,
			want: want{
				contentType: "application/json",
				code:        http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten/batch", strings.NewReader(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, uuid.NewString())
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", tt.contentType)
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.code, res.Code)
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, res.Body.String())
			}
		})
	}
}

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)
	return r
}
