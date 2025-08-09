package main

import (
	"compress/gzip"
	"io"
	chess "main/pkg"
	"net/http"
	"strings"

	"github.com/syumai/workers"
)

type gzipFuncResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipFuncResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Compress(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipFuncResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}

func main() {
	http.HandleFunc("GET /", chess.FuncHandler)
	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	workers.Serve(nil)
}
