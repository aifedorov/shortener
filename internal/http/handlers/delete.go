package handlers

import (
	"errors"
	"net/http"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/repository"
)

func NewDeleteHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		aliases, err := decodeAliasesRequest(r)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !validateAliases(aliases) {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userID, err := getUseID(r)
		if errors.Is(err, repository.ErrURLDeleted) {
			rw.WriteHeader(http.StatusGone)
			return
		}
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = repo.DeleteBatch(userID, aliases)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("sending HTTP 202 response")
		rw.WriteHeader(http.StatusAccepted)
	}
}
