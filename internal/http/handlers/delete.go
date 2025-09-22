package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
)

func NewDeleteHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		aliases, err := decodeAliasesRequest(r)
		if err != nil || len(aliases) == 0 {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userID, err := getUserID(r)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		go func(userID string, aliases []string) {
			err = repo.DeleteBatch(userID, aliases)
			if err != nil {
				logger.Log.Error("failed to delete urls", zap.Error(err))
				return
			}
		}(userID, aliases)

		logger.Log.Debug("sending HTTP 202 response")
		rw.WriteHeader(http.StatusAccepted)
	}
}
