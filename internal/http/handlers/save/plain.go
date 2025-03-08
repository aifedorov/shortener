package save

import (
	"errors"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/validate"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func NewSavePlainTextHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")

		body, readErr := io.ReadAll(req.Body)
		if readErr != nil {
			logger.Log.Error("failed to read request body", zap.Error(readErr))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		url := string(body)
		if err := urlChecker.CheckURL(url); err != nil {
			logger.Log.Error("invalid url", zap.String("url", url))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resURL, saveErr := repo.Store(config.BaseURL, url)
		if errors.Is(saveErr, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusOK)
			_, writeErr := rw.Write([]byte(resURL))
			if writeErr != nil {
				logger.Log.Error("Failed to write response", zap.Error(writeErr))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		rw.WriteHeader(http.StatusCreated)
		_, writeErr := rw.Write([]byte(resURL))
		if writeErr != nil {
			logger.Log.Error("Failed to write response", zap.Error(writeErr))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
