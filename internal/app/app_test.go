package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestServer_methodGetHandler(t *testing.T) {
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
			server: NewServer(),
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
				pathToURL: map[string]string{
					"1": "https://google.com",
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
				pathToURL: map[string]string{
					"1": "https://google.com",
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

func TestServer_methodPostHandler(t *testing.T) {
	type want struct {
		expectedContentType string
		expectedCode        int
		expectedBody        string
	}
	tests := []struct {
		name        string
		server      *Server
		method      string
		requestBody string
		want        want
	}{
		{
			name:   "Post method without parameters",
			server: NewServer(),
			method: http.MethodPost,
			want: want{
				expectedContentType: "text/plain",
				expectedCode:        http.StatusCreated,
			},
		},
		{
			name:        "Post method with url",
			server:      NewServer(),
			method:      http.MethodPost,
			requestBody: `https://google.com`,
			want: want{
				expectedContentType: "text/plain",
				expectedCode:        http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.server.mountHandlers()
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			res := executeRequest(req, tt.server)

			assert.Equal(t, tt.want.expectedCode, res.Code)
			assert.Equal(t, tt.want.expectedContentType, res.Header().Get("Content-Type"))

			if tt.want.expectedBody != "" {
				assert.Equal(t, tt.want.expectedBody, res.Body.String())
			}
		})
	}
}

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	s.router.ServeHTTP(r, req)
	return r
}
