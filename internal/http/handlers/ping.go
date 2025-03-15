package handlers

import (
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"net/http"

	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/repository"
)

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
