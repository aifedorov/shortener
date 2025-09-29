package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData holds response information for logging purposes.
	responseData struct {
		// status is the HTTP status code.
		status int
		// size is the response body size in bytes.
		size int
		// body is the response body content.
		body []byte
	}

	// loggingResponseWriter wraps http.ResponseWriter to capture response data for logging.
	loggingResponseWriter struct {
		http.ResponseWriter
		// responseData holds the captured response information.
		responseData *responseData
	}
)

// Write writes data to the underlying response writer and captures it for logging.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.body = append(r.responseData.body, b...)
	return size, err
}

// WriteHeader writes the HTTP status code and captures it for logging.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Log is the global logger instance used throughout the application.
var Log = zap.NewNop()

// Initialize sets up the logger with the specified log level.
// It configures the logger for development mode with the given level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		Log.Error("logger: failed to parse level", zap.Error(err))
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		Log.Error("logger: failed to build", zap.Error(err))
		return err
	}

	Log = zl
	return nil
}

// RequestLogger provides HTTP request logging middleware.
// It logs incoming requests with method, URL, headers, and processing duration.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		Log.Info("HTTP request ==>",
			zap.String("method", r.Method),
			zap.String("URL", r.URL.String()),
			zap.Any("headers", r.Header),
			zap.Duration("duration", duration),
		)
	})
}

// ResponseLogger provides HTTP response logging middleware.
// It logs outgoing responses with status code, headers, body, size, and processing duration.
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

		Log.Info("HTTP response <==",
			zap.Int("status", rd.status),
			zap.Any("headers", r.Header),
			zap.ByteString("body", rd.body),
			zap.Int("size", rd.size),
			zap.Duration("duration", duration),
		)
	})
}
