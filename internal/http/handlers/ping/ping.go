package ping

import (
	"net/http"

	"github.com/aifedorov/shortener/internal/repository"
	"github.com/aifedorov/shortener/pkg/logger"
)

func NewPingHandler(repo repository.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if pgr, ok := repo.(*repository.PostgresRepository); ok {
			err := pgr.Connect(req.Context())
			if err != nil {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		logger.Log.Info("ping: use file or in memory storage")
		rw.WriteHeader(http.StatusOK)
	}
}
