package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		Log.Info("HTTP request complete",
			zap.String("method", r.Method),
			zap.String("host", r.Host),
			zap.String("path", r.URL.String()),
			zap.Duration("duration", duration),
		)
	})
}

func ResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rd := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   rd,
		}

		start := time.Now()
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)

		Log.Info("HTTP response complete",
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
			zap.Duration("duration", duration),
		)
	})
}
