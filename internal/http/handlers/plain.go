package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/validate"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
)

func NewSavePlainTextHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")

		userID, err := getUserID(r)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		logger.Log.Debug("reading request body")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("failed to read request body", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		logger.Log.Debug("checking original url", zap.String("original_url", string(body)))
		oURL := string(body)
		if err := urlChecker.CheckURL(oURL); err != nil {
			logger.Log.Error("invalid original url", zap.String("original_url", oURL))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if oURL == "" {
			oURL = config.BaseURL
		}

		logger.Log.Debug("saving original url", zap.String("original_url", oURL))
		resURL, err := repo.Store(userID, config.BaseURL, oURL)
		var cErr *repository.ConflictError
		if errors.As(err, &cErr) {
			logger.Log.Debug("sending HTTP 409 response")
			rw.WriteHeader(http.StatusConflict)

			_, err := rw.Write([]byte(cErr.ShortURL))
			if err != nil {
				logger.Log.Error("failed to write response", zap.Error(err))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		if err != nil {
			logger.Log.Error("failed to save original url", zap.String("original_url", oURL), zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logger.Log.Debug("store updated", zap.String("short_url", resURL), zap.String("original_url", oURL))

		logger.Log.Debug("sending HTTP 201 response")
		rw.WriteHeader(http.StatusCreated)
		_, writeErr := rw.Write([]byte(resURL))
		if writeErr != nil {
			logger.Log.Error("Failed to write response", zap.Error(writeErr))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
