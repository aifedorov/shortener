package server

import (
	"fmt"
	"github.com/aifedorov/shortener/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aifedorov/shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestServer_redirect(t *testing.T) {
	t.Parallel()

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
			server: NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server: &Server{
				router: chi.NewRouter(),
				store: &storage.MemoryStorage{
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			},
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
			server: &Server{
				router: chi.NewRouter(),
				store: &storage.MemoryStorage{
					PathToURL: map[string]string{
						"1": "https://google.com",
					},
				},
			},
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
			method:      http.MethodPost,
			contentType: "text/plain",
			want: want{
				contentType: "text/plain",
				code:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with valid url",
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
				&storage.MemoryStorage{
					PathToURL: map[string]string{"BQRvJsg-jIg": "https://google.com"},
				},
			),
			method:      http.MethodPost,
			contentType: "text/plain",
			requestBody: `https://google.com`,
			want: want{
				contentType: "text/plain",
				code:        http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server:      NewServer(config.NewConfig(), storage.NewMemoryStorage()),
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
			server: &Server{
				router: chi.NewRouter(),
				store: &storage.MemoryStorage{
					PathToURL: map[string]string{
						"BQRvJsg-jIg": "https://google.com",
					},
				},
				config: config.NewConfig(),
			},
			method:      http.MethodPost,
			contentType: "application/json",
			requestBody: `{"url": "https://google.com"}`,
			want: want{
				contentType: "application/json",
				code:        http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.requestBody))
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
