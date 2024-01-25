package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	num, err := w.Writer.Write(b)
	return num, err
}

func WithCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader

		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				_, werr := io.WriteString(w, err.Error())
				if werr != nil {
					log.Error().Err(werr).Msg("failed to write error")
				}
				return
			}
			reader = gz
			defer gz.Close()

			r.Body = io.NopCloser(reader)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			_, werr := io.WriteString(w, err.Error())
			if werr != nil {
				log.Error().Err(werr).Msg("failed to write error")
			}
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
