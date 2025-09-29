package handlers

import (
	"errors"
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/repository"
)

// URLResponse represents a URL entry in the user's URL list response.
// Used for returning user's URLs in the /api/user/urls endpoint.
type URLResponse struct {
	// ShortURL is the generated short URL.
	ShortURL string `json:"short_url"`
	// OriginalURL is the original URL that was shortened.
	OriginalURL string `json:"original_url"`
}

// NewURLsHandler creates a new HTTP handler for retrieving all URLs belonging to a user.
// This handler requires user authentication. If the user is not authenticated, a cookie will be created for them.
// It returns a handler function that responds with a JSON array of user's URLs.
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
