package compress

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"slices"
	"sync"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"go.uber.org/zap"
)

// writerPool is a pool of gzip writers for efficient memory reuse.
var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// compressWriter wraps an http.ResponseWriter with gzip compression capabilities.
type compressWriter struct {
	// w is the underlying HTTP response writer.
	w http.ResponseWriter
	// zw is the gzip writer for compression.
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter instance.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := writerPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

// compressReader wraps an io.ReadCloser with gzip decompression capabilities.
type compressReader struct {
	// r is the underlying reader.
	r io.ReadCloser
	// zr is the gzip reader for decompression.
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader instance.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Header returns the HTTP header map.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes compressed data to the underlying writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes the HTTP status code and sets Content-Encoding header for successful responses.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer and returns it to the pool.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	c.zw.Reset(io.Discard)
	writerPool.Put(c.zw)
	return err
}

// Read reads decompressed data from the underlying reader.
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the underlying reader and the gzip reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware provides gzip compression/decompression middleware for HTTP handlers.
// It compresses responses when the client supports gzip and decompresses gzip-encoded requests.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		supportsGzip := slices.Contains(r.Header.Values("Accept-Encoding"), "gzip")
		if supportsGzip {
			logger.Log.Debug("compress: compressing response body")
			cw := newCompressWriter(w)
			defer func() {
				err := cw.Close()
				if err != nil {
					logger.Log.Error("compress: failed to close gzip writer", zap.Error(err))
				}
			}()

			ow = cw
		}

		sendsGzip := slices.Contains(r.Header.Values("Content-Encoding"), "gzip")
		if sendsGzip {
			logger.Log.Debug("compress: decompressing request body")
			cr, err := newCompressReader(r.Body)
			if errors.Is(err, io.EOF) {
				next.ServeHTTP(ow, r)
				return
			}

			if err != nil {
				logger.Log.Error("compress: failed to decompress request body", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				err := cr.Close()
				if err != nil {
					logger.Log.Error("compress: failed to close gzip reader", zap.Error(err))
				}
			}()
		}

		next.ServeHTTP(ow, r)
	})
}
