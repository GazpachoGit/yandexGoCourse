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
	rContentType := strings.Join(w.Header().Values("Content-Type"), ",")
	if checkContentType(rContentType) {
		return w.Writer.Write(b)
	} else {
		return w.Write(b)
	}
}

type Compressor struct {
	gz *gzip.Writer
}

func (comp *Compressor) CompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rEncoding := strings.Join(r.Header.Values("Accept-Encoding"), ",")
		if strings.Contains(rEncoding, "gzip") {
			if comp.gz == nil {
				gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				comp.gz = gz
			} else {
				comp.gz.Reset(w)
			}
			defer comp.gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: comp.gz}, r)
			return
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
