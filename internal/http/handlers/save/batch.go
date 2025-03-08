package save

import (
	"encoding/json"
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
		var reqURLs []BatchRequest

		if err := json.NewDecoder(req.Body).Decode(&reqURLs); err != nil {
			logger.Log.Error("failed to decode request", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		logger.Log.Debug("validating requested urls", zap.Int("count", len(reqURLs)))
		var urls = make([]repository.URLInput, len(reqURLs))
		for i, reqBodyURL := range reqURLs {
			if err := urlChecker.CheckURL(reqBodyURL.OriginalURL); err != nil {
				logger.Log.Error("invalid url parameter in request", zap.String("request", reqBodyURL.String()))
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			urls[i] = repository.URLInput{
				CID:         reqBodyURL.CID,
				OriginalURL: reqBodyURL.OriginalURL,
			}
		}

		res, err := repo.StoreBatch(config.BaseURL, urls)
		if err != nil {
			logger.Log.Error("failed to store batch", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("sending HTTP 201 response")
		rw.WriteHeader(http.StatusCreated)
		if err := encodeBatchResponse(rw, res); err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
