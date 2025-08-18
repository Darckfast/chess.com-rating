package main

import (
	chess "main/pkg"
	"net/http"

	"github.com/syumai/workers/cloudflare/fetch"
)

func main() {
	http.HandleFunc("GET /", chess.FuncHandler)
	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	fetch.ServeNonBlock(nil)

	<-make(chan struct{})
}
