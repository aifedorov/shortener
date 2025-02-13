package redirect

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/internal/storage"
)

func NewRedirectHandler(storage storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		short := chi.URLParam(req, "shortURL")
		target, err := storage.GetURL(short)
		if err != nil {
			logger.Log.Debug("short url is not found", zap.String("short url", short))
			http.NotFound(res, req)
			return
		}

		http.Redirect(res, req, target, http.StatusTemporaryRedirect)
	}
}
