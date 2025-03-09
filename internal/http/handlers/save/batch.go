package save

import (
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/validate"
	"go.uber.org/zap"
	"net/http"
)

func NewSaveJSONBatchHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		logger.Log.Debug("decoding request body")
		reqURLs, err := decodeRequest(req)
		if err != nil {
			logger.Log.Error("failed to decode request", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		logger.Log.Debug("validating requested urls", zap.Int("count", len(reqURLs)))
		urls, err := validateURLs(reqURLs, urlChecker)
		if err != nil {
			logger.Log.Error("invalid url parameter in request")
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		res, err := repo.StoreBatch(config.BaseURL, urls)
		if err != nil {
			logger.Log.Error("failed to store batch", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("sending HTTP 201 response")
		rw.WriteHeader(http.StatusCreated)
		logger.Log.Debug("encoding response body", zap.Any("body", res))
		if err := encodeBatchResponse(rw, res); err != nil {
			logger.Log.Error("failed to encode batch response", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
