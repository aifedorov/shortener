package ping

import (
	"context"
	"database/sql"
	"github.com/aifedorov/shortener/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func NewPingHandler(dsn string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			logger.Log.Error("failed to open database", zap.Error(err))
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer func() {
			err := db.Close()
			if err != nil {
				logger.Log.Error("failed to close database", zap.Error(err))
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}
