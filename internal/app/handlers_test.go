package app

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_methodGetHandler(t *testing.T) {
	type want struct {
		expectedContentType string
		expectedCode        int
		expectedLocation    string
		expectedBody        string
	}
	tests := []struct {
		name    string
		server  *Server
		method  string
		request string
		want    want
	}{
		{
			name:    "Get method without id",
			server:  NewServer(),
			method:  http.MethodGet,
			request: ``,
			want: want{
				expectedContentType: "text/plain; charset=utf-8",
				expectedCode:        http.StatusBadRequest,
				expectedBody:        fmt.Sprintf("%s\n", ErrShortURLMissing.Error()),
			},
		},
		{
			name: "Get method with existing id",
			server: &Server{
				mux: http.NewServeMux(),
				pathToURL: map[string]string{
					"1": "https://google.com",
				},
			},
			method:  http.MethodGet,
			request: `1`,
			want: want{
				expectedContentType: "text/html; charset=utf-8",
				expectedCode:        http.StatusTemporaryRedirect,
				expectedLocation:    "https://google.com",
			},
		},
		{
			name: "Get method with not existing id",
			server: &Server{
				mux: http.NewServeMux(),
				pathToURL: map[string]string{
					"1": "https://google.com",
				},
			},
			method:  http.MethodGet,
			request: `2`,
			want: want{
				expectedContentType: "text/plain; charset=utf-8",
				expectedCode:        http.StatusNotFound,
				expectedBody:        "404 page not found\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, fmt.Sprintf("/%s", tt.request), nil)
			w := httptest.NewRecorder()

			tt.server.methodGetHandler(w, r)

			res := w.Result()

			assert.Equal(t, tt.want.expectedCode, res.StatusCode)
			assert.Equal(t, tt.want.expectedContentType, res.Header.Get("Content-Type"))

			defer func() {
				err := res.Body.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if tt.want.expectedLocation != "" {
				assert.Equal(t, tt.want.expectedLocation, res.Header.Get("Location"))
			}

			if tt.want.expectedBody != "" {
				assert.Equal(t, tt.want.expectedBody, string(resBody))
			}
		})
	}
}
