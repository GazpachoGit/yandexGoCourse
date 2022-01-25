package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func CompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rEncoding := strings.Join(r.Header.Values("Accept-Encoding"), ",")
		rContentType := strings.Join(r.Header.Values("Content-Type"), ",")

		if strings.Contains(rEncoding, "gzip") && checkContentType(rContentType) {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
		}
		next.ServeHTTP(w, r)
	})
}

func checkContentType(rContentType string) bool {
	contentTypes := []string{
		"application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml",
	}
	for _, ct := range contentTypes {
		if strings.Contains(rContentType, ct) {
			return true
		}
	}
	return false
}
