package chess

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
)

const (
	CHESS_CALLBACK_URL = "https://www.chess.com/callback/member/stats/"
)

var Log *slog.Logger

func InitLoggger(r *http.Request) {
	c := fetch.NewClient()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cloudflare.Getenv("SENTRY_DSN"),
		EnableLogs:  true,
		HTTPClient:  c.HTTPClient(fetch.RedirectModeError),
		Environment: cloudflare.Getenv("WORKER_ENV"),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	requestIp := r.Header.Get("X-Forwarded-For")
	connectingIp := r.Header.Get("CF-Connecting-IP")
	if requestIp != "" && connectingIp != "" {
		requestIp += ","
	}
	requestIp += connectingIp
	handler := sentryslog.Option{
		LogLevel: []slog.Level{slog.LevelWarn, slog.LevelInfo, slog.LevelError},
	}.NewSentryHandler(r.Context())

	if cloudflare.Getenv("WORKER_ENV") == "dev" {
		prettyHandler := log.New(os.Stdout)
		Log = slog.New(slogmulti.Fanout(handler, prettyHandler))
		Log = Log.With(
			slog.Group("http",
				slog.String("method", r.Method),
				slog.String("route", r.URL.Path),
			),
		)
	} else {
		Log = slog.New(handler)
		Log = Log.With(
			slog.Group("http",
				slog.String("method", r.Method),
				slog.String("route", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("user-agent", r.UserAgent()),
			),
		).
			With("environment", cloudflare.Getenv("WORKER_ENV")).
			With("service", "chess.com-rating")
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	InitLoggger(r)

	defer func() {
		sentry.Flush(1 * time.Second)
	}()

	w.Header().Set("Content-Type", "text/plain")
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	ctx := r.Context()
	username := r.URL.Query().Get("username")
	message := r.URL.Query().Get("message")
	message, _ = url.QueryUnescape(message)
	username, _ = url.QueryUnescape(username)
	username = strings.TrimSpace(username)

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
