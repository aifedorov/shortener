package save

import (
	"encoding/json"
	"errors"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/validate"
	"go.uber.org/zap"
	"net/http"
)

func NewSaveJSONHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		logger.Log.Debug("decoding request body")
		var reqBody Request
		if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
			logger.Log.Error("failed to decode request", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := urlChecker.CheckURL(reqBody.URL); err != nil {
			logger.Log.Error("invalid url parameter in request", zap.String("url", reqBody.URL))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resURL, saveErr := repo.Store(config.BaseURL, reqBody.URL)
		if errors.Is(saveErr, repository.ErrURLExists) {
			logger.Log.Debug("sending HTTP 200 response")
			rw.WriteHeader(http.StatusOK)

			logger.Log.Debug("encoding response", zap.Any("response", rw))
			if err := encodeResponse(rw, resURL); err != nil {
				logger.Log.Error("failed to encode response", zap.Error(err))
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}

		logger.Log.Debug("sending HTTP 201 response")
		rw.WriteHeader(http.StatusCreated)
		if err := encodeResponse(rw, resURL); err != nil {
			logger.Log.Error("failed to send HTTP 201 response", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
