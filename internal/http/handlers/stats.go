package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
	"go.uber.org/zap"
)

// NewStatsHandler creates a new HTTP handler for retrieving service statistics.
// This endpoint requires the client IP to be within the trusted subnet.
// It returns the total number of shortened URLs and users in the service.
// Only the GET method is allowed.
func NewStatsHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rw.Header().Set("Content-Type", "application/json")

		stats, err := repo.GetStats()
		if err != nil {
			logger.Log.Error("stats: failed to get stats", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response := StatsResponse{
			URLs:  stats.TotalURLs,
			Users: stats.TotalUsers,
		}

		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(response); err != nil {
			logger.Log.Error("stats: failed to encode response", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
