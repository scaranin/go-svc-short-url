package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter implements http.ResponseWriter and transparently compresses data
// written to it using gzip encoding. It automatically sets the appropriate
// Content-Encoding header when the status code indicates success (<300).
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter instance wrapping the provided
// http.ResponseWriter and initializing a new gzip.Writer.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the header map from the original ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses the data using gzip before writing to the underlying ResponseWriter.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader sets the status code and adds Content-Encoding: gzip header
// for successful responses (status code < 300).
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close flushes any pending compressed data and closes the gzip writer.
// This should be called when finished with the writer to ensure all data is sent.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader implements io.ReadCloser and transparently decompresses
// gzip-encoded request bodies.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader instance wrapping the provided
// io.ReadCloser and initializing a new gzip.Reader.
// Returns error if the gzip reader cannot be initialized.
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

// Read decompresses data from the underlying gzip reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the gzip reader and the original reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware provides HTTP middleware that transparently handles gzip
// compression/decompression. It:
// - Decompresses gzip-encoded request bodies when Content-Encoding: gzip is present
// - Compresses responses when Accept-Encoding: gzip is present in the request
// - Sets appropriate headers automatically
//
// Usage:
//
//	router.Use(GzipMiddleware)
func GzipMiddleware(h http.Handler) http.Handler {
	gzipFunc := func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			ow.Header().Set("Content-Encoding", "gzip")

			cw := newCompressWriter(w)
			ow = cw

			defer cw.Close()
		}
		h.ServeHTTP(ow, r)

	}
	return http.HandlerFunc(gzipFunc)
}
