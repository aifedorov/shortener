package handlers

import (
	"errors"
	"net/http"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/validate"
)

func NewSaveJSONBatchHandler(config *config.Config, repo repository.Repository, urlChecker validate.URLChecker) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		reqURLs, err := decodeBatchRequest(r)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		urls, err := validateURLs(reqURLs, urlChecker)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userID, err := getUseID(r)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res, err := repo.StoreBatch(userID, config.BaseURL, urls)
		var cErr *repository.ConflictError
		if errors.As(err, &cErr) {
			logger.Log.Debug("sending HTTP 409 response")
			rw.WriteHeader(http.StatusConflict)

			if err := encodeResponse(rw, cErr.ShortURL); err != nil {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		if err != nil {
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
