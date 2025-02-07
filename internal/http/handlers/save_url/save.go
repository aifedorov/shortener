package save_url

import (
	"errors"
	"io"
	"net/http"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/storage"
)

var (
	ErrURLMissing = errors.New("URL is missing")
)

func NewURLSaveHandler(config *config.Config, store storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")

		body, readErr := io.ReadAll(req.Body)
		if errors.Is(readErr, io.EOF) {
			http.Error(res, ErrURLMissing.Error(), http.StatusBadRequest)
			return
		}
		if readErr != nil {
			http.Error(res, readErr.Error(), http.StatusBadRequest)
			return
		}

		resURL, saveErr := store.SaveURL(config.ShortBaseURL, string(body))
		if errors.Is(saveErr, storage.ErrURLExists) {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(resURL))
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusCreated)
		_, writeErr := res.Write([]byte(resURL))
		if writeErr != nil {
			http.Error(res, writeErr.Error(), http.StatusInternalServerError)
			return
		}
	}
}
