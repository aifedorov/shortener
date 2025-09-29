package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aifedorov/shortener/internal/mocks"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewRedirectHandler(t *testing.T) {
	tests := []struct {
		name             string
		shortURL         string
		getResult        string
		getError         error
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "successful redirect",
			shortURL:         "abc123",
			getResult:        "https://example.com",
			getError:         nil,
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:             "short URL not found",
			shortURL:         "nonexistent",
			getResult:        "",
			getError:         repository.ErrShortURLNotFound,
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
		{
			name:             "URL deleted",
			shortURL:         "deleted123",
			getResult:        "",
			getError:         repository.ErrURLDeleted,
			expectedStatus:   http.StatusGone,
			expectedLocation: "",
		},
		{
			name:             "repository error",
			shortURL:         "error123",
			getResult:        "",
			getError:         errors.New("database error"),
			expectedStatus:   http.StatusInternalServerError,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().Get(tt.shortURL).Return(tt.getResult, tt.getError)

			handler := NewRedirectHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)

			r := chi.NewRouter()
			r.Get("/{shortURL}", handler)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, rr.Header().Get("Location"))
			}
		})
	}
}

func TestNewRedirectHandler_ChiURLParam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Get("test123").Return("https://example.com", nil)

	handler := NewRedirectHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/test123", nil)

	r := chi.NewRouter()
	r.Get("/{shortURL}", handler)

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "https://example.com", rr.Header().Get("Location"))
}

func TestNewRedirectHandler_DirectCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().Get("direct123").Return("https://example.com", nil)

	handler := NewRedirectHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/direct123", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("shortURL", "direct123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler(rr, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "https://example.com", rr.Header().Get("Location"))
}
