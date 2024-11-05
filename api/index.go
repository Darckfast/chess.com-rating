package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type ChessComResponse struct {
	ChessRapid struct {
		Last struct {
			Rating int `json:"rating"`
			Date   int `json:"date"`
			Rd     int `json:"rd"`
		} `json:"last"`
	} `json:"chess_rapid"`
	Tactics struct {
		Highest struct {
			Rating int `json:"rating"`
			Date   int `json:"date"`
		} `json:"highest"`
	} `json:"tactics"`
}

const (
	CHESS_API_URL = "https://api.chess.com/pub/player/"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func Handler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Incoming request")
	username := r.URL.Query().Get("username")
	username, _ = url.QueryUnescape(username)
	username = strings.TrimSpace(username)

	if username == "" {
		w.WriteHeader(400)

		fmt.Fprint(w, "no username queried")
		return
	}

	req, _ := http.NewRequest("GET", CHESS_API_URL+url.QueryEscape(username)+"/stats", nil)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Error("error requesting chess.com api", "error", err.Error())

		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	if res.StatusCode != http.StatusOK {
		logger.Error("chess.com api returned error", "status", res.StatusCode)

		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	var chessRes ChessComResponse
	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&chessRes)

	fmt.Fprint(w, chessRes.ChessRapid.Last.Rating)
}
