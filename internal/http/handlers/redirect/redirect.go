package redirect

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aifedorov/shortener/internal/storage"
)

func NewRedirectHandler(storage storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		short := chi.URLParam(req, "shortURL")
		target, err := storage.GetURL(short)
		if err != nil {
			http.NotFound(res, req)
			return
		}

		http.Redirect(res, req, target, http.StatusTemporaryRedirect)
	}
}
