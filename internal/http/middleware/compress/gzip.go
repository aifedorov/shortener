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

var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := writerPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	c.zw.Reset(io.Discard)
	writerPool.Put(c.zw)
	return err
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

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
