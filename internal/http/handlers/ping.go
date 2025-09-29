package handlers

import (
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"

	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/repository"
)

// NewPingHandler creates a new HTTP handler for health check operations.
// This handler is available to all users (no authentication required).
// It tests the connection to the underlying storage and returns appropriate status codes.
func NewPingHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		err := repo.Ping()
		if err != nil {
			logger.Log.Error("ping failed", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("sending HTTP 200 code", zap.String("url", req.URL.String()))
		rw.WriteHeader(http.StatusOK)
	}
}
