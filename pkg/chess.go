package chess

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/syumai/workers/cloudflare/fetch"
)

const (
	CHESS_CALLBACK_URL = "https://www.chess.com/callback/member/stats/"
)

var Log = slog.New(NewHandler(os.Stdout))

func FuncHandler(w http.ResponseWriter, r *http.Request) {
	wg := SetupContext(r)

	defer wg.Wait()

	w.Header().Set("Content-Type", "text/plain")
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	ctx := r.Context()
	username := r.URL.Query().Get("username")
	message := r.URL.Query().Get("message")
	username = strings.TrimSpace(username)
	message, _ = url.QueryUnescape(message)
	username, _ = url.QueryUnescape(username)

	if username == "" {
		w.WriteHeader(400)
		Log.WarnContext(ctx, "username is required")
		fmt.Fprint(w, "username is required")
		return
	}

	client := fetch.NewClient()
	req, _ := fetch.NewRequest(r.Context(), "GET", CHESS_CALLBACK_URL+url.QueryEscape(username), nil)
	res, err := client.Do(req, nil)
	if err != nil {
		Log.ErrorContext(ctx, "error requesting chess.com", slog.Any("error", err))
		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	if res.StatusCode != http.StatusOK {
		Log.ErrorContext(ctx, "chess.com returned error", slog.Int("status", res.StatusCode))
		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	var chess ChessStats
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&chess)
	for _, stat := range chess.Stats {
		if strings.Contains(message, stat.Key) {
			message = strings.ReplaceAll(message, "="+stat.Key, string(stat.Stats.Rating))
		}
	}

	Log.InfoContext(ctx, "request completed", "status", 200)
	fmt.Fprint(w, message)
}
