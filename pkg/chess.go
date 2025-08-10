package chess

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Darckfast/axiom-log-this-go/pkg/logthis"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
)

const (
	CHESS_COM_URL = "https://www.chess.com/callback/member/stats/"
)

var Log = logthis.NewLogger(&logthis.NewHandlerArgs{
	Out:         os.Stdout,
	ServiceName: "chess.com-rating",
	AxiomApiKey: cloudflare.Getenv("AXIOM_API_KEY"),
	Transport:   fetch.NewClient().HTTPClient(fetch.RedirectModeFollow).Transport,
})

func FuncHandler(w http.ResponseWriter, r *http.Request) {
	wg, r, _ := logthis.FromRequest(r)

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
	req, _ := fetch.NewRequest(r.Context(), "GET", CHESS_COM_URL+url.QueryEscape(username), nil)
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

	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "Deny")
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	Log.InfoContext(ctx, "response", "status", 200)
	fmt.Fprint(w, message)
}
