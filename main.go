package main

import (
	"net/http"

	chess "main/api/v1"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", chess.Handler)
	workers.Serve(nil)
}
