//go:build js && wasm

package chess

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Darckfast/workers-go/cloudflare/fetch"
)

const (
	CHESS_COM_URL = "https://www.chess.com/callback/member/stats/"
)

var client = fetch.Client{
	Timeout: 3 * time.Second,
}

var Log = NewLogger(&NewHandlerArgs{
	out:         os.Stdout,
	serviceName: "chess.com-ratings",
	axiomApiKey: os.Getenv("AXIOM_API_KEY"),
})

func FuncHandler(w http.ResponseWriter, r *http.Request) {
	wg, r := FromContext(r)

	if wg != nil {
		defer wg.Wait()
	}

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
		w.WriteHeader(200)
		Log.WarnContext(ctx, "username is required")
		w.Write([]byte("username is required"))
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), "GET", CHESS_COM_URL+url.QueryEscape(username), nil)
	res, err := client.Do(req)
	if err != nil {
		Log.ErrorContext(ctx, "error requesting chess.com", slog.Any("error", err))
		w.Write([]byte("ops, something went wrong"))
		return
	}

	if res.StatusCode != http.StatusOK {
		Log.ErrorContext(ctx, "chess.com returned error", slog.Int("status", res.StatusCode))
		w.Write([]byte("ops, something went wrong"))
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

	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "Deny")
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	Log.InfoContext(ctx, "response", "status", 200)
	w.Write([]byte(message))
}
