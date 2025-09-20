package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello World"))
	}))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := "Hello World"
	responseBody := "Hello World"

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEqual(t, responseBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, responseBody, string(b))
	})
}

func BenchmarkGzipMiddleware(b *testing.B) {
	payload := []byte(`{"url":"https://example.com"}`)

	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))

	b.Run("with_compression", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})

	b.Run("without_compression", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkCompressWriter(b *testing.B) {
	data := []byte(`{"short_url":"abc","original_url":"https://example.com"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		cw := newCompressWriter(w)
		cw.Write(data)
		cw.Close()
	}
}

func BenchmarkCompressReader(b *testing.B) {
	data := []byte(`{"urls":[{"short_url":"abc123","original_url":"https://example.com"}]`)

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write(data)
	zw.Close()

	compressedData := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(compressedData)
		cr, err := newCompressReader(io.NopCloser(reader))
		if err != nil {
			b.Fatal(err)
		}

		_, err = io.ReadAll(cr)
		if err != nil {
			b.Fatal(err)
		}
		cr.Close()
	}
}
