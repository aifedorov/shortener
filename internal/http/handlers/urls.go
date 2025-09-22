package handlers

import (
	"errors"
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
)

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURLsHandler(cfg *config.Config, repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		userID, err := getUserID(r)
		if err != nil {
			logger.Log.Error("error getting user id", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		logger.Log.Debug("fetching urls for user_id", zap.String("user_id", userID))
		urls, err := repo.GetAll(userID, cfg.BaseURL)
		if errors.Is(err, repository.ErrUserHasNoData) {
			logger.Log.Info("user don't have any urls", zap.String("user_id", userID))
			http.Error(rw, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}
		if err != nil {
			logger.Log.Error("error fetching urls", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		err = encodeURLsResponse(rw, urls)
		if err != nil {
			logger.Log.Error("error encoding urls", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
