package handlers

import (
	"errors"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/repository"
)

func NewRedirectHandler(repo repository.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "shortURL")
		target, err := repo.Get(shortURL)
		if errors.Is(err, repository.ErrShortURLNotFound) {
			logger.Log.Info("short url not found", zap.String("alias", shortURL))
			http.NotFound(res, req)
			return
		}
		if err != nil {
			logger.Log.Error("failed to get short url", zap.String("short_url", shortURL), zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Info("redirecting to url", zap.String("alias", shortURL), zap.String("url", target))
		http.Redirect(res, req, target, http.StatusTemporaryRedirect)
	}
}
