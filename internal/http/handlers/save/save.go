package save

import (
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/internal/storage"
	"github.com/aifedorov/shortener/lib/validate"
)

func NewURLSaveHandler(config *config.Config, store storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")

		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			logger.Log.Debug("got request with empty body")
			http.Error(res, readErr.Error(), http.StatusBadRequest)
			return
		}

		url := string(body)
		if err := validate.CheckURL(url); err != nil {
			logger.Log.Debug("got request with invalid input url", zap.String("url", url))
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		resURL, saveErr := store.SaveURL(config.ShortBaseURL, url)
		if errors.Is(saveErr, storage.ErrURLExists) {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(resURL))
			if err != nil {
				logger.Log.Debug("failed to write response url", zap.String("response url", resURL))
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		res.WriteHeader(http.StatusCreated)
		_, writeErr := res.Write([]byte(resURL))
		if writeErr != nil {
			logger.Log.Debug("failed to write response url", zap.String("response url", resURL))
			http.Error(res, writeErr.Error(), http.StatusInternalServerError)
			return
		}
	}
}
