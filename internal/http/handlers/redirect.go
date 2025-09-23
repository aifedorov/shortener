package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
)

// NewRedirectHandler creates a new HTTP handler for redirecting short URLs to their original URLs.
// This handler is available to all users (no authentication required).
// It returns a handler function that performs HTTP redirects or returns appropriate error responses.
func NewRedirectHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "shortURL")
		target, err := repo.Get(shortURL)
		if errors.Is(err, repository.ErrShortURLNotFound) {
			logger.Log.Info("redirect: short url not found", zap.String("alias", shortURL))
			http.NotFound(rw, r)
			return
		}
		if errors.Is(err, repository.ErrURLDeleted) {
			logger.Log.Info("redirect: url deleted", zap.String("alias", shortURL))
			http.Error(rw, http.StatusText(http.StatusGone), http.StatusGone)
		}
		if err != nil {
			logger.Log.Error("redirect: failed to get short url", zap.String("short_url", shortURL), zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Info("redirect: redirecting to url", zap.String("alias", shortURL), zap.String("url", target))
		http.Redirect(rw, r, target, http.StatusTemporaryRedirect)
	}
}
