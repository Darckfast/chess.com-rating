package main

import (
	chess "main/pkg"
	"net/http"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", chess.FuncHandler)
	workers.Serve(nil)
}
