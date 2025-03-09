package redirect

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
)

func NewRedirectHandler(repo repository.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "shortURL")
		target, err := repo.Get(shortURL)
		if err != nil {
			logger.Log.Error("shortener shortener not found", zap.String("shortURL", shortURL))
			http.NotFound(res, req)
			return
		}

		logger.Log.Debug("redirecting to url", zap.String("url", target))
		http.Redirect(res, req, target, http.StatusTemporaryRedirect)
	}
}
