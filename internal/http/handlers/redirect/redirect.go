package redirect

import (
	"github.com/aifedorov/shortener/pkg/logger"
	"go.uber.org/zap"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aifedorov/shortener/internal/repository"
)

func NewRedirectHandler(repo repository.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		short := chi.URLParam(req, "shortURL")
		target, err := repo.Get(short)
		if err != nil {
			logger.Log.Debug("short url is not found", zap.String("short url", short))
			http.NotFound(res, req)
			return
		}

		http.Redirect(res, req, target, http.StatusTemporaryRedirect)
	}
}
