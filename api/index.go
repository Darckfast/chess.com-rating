package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"main/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
)

const (
	CHESS_CALLBACK_URL = "https://www.chess.com/callback/member/stats/"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Namespace:   r.URL.Path,
		ApiKey:      os.Getenv("BASELIME_API_KEY"),
		ServiceName: os.Getenv("VERCEL_GIT_REPO_SLUG"),
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	logger.InfoContext(ctx, "Processing request")

	username := r.URL.Query().Get("username")
	message := r.URL.Query().Get("message")
	message, _ = url.QueryUnescape(message)
	keys := r.URL.Query()["key"]

	username, _ = url.QueryUnescape(username)
	username = strings.TrimSpace(username)

	if username == "" {
		w.WriteHeader(400)
		logger.WarnContext(ctx, "username is required")
		fmt.Fprint(w, "username is required")
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(400)
		logger.WarnContext(ctx, "stats keys are required")
		fmt.Fprint(w, "stats keys are required")
		return
	}

	req, _ := http.NewRequest("GET", CHESS_CALLBACK_URL+url.QueryEscape(username), nil)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		logger.ErrorContext(ctx, "error requesting chess.com api", "error", err.Error())

		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	if res.StatusCode != http.StatusOK {
		logger.ErrorContext(ctx, "chess.com api returned error", "status", res.StatusCode)

		fmt.Fprint(w, "ops, something went wrong")
		return
	}

	var chessRes utils.MemberCallback
	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&chessRes)

	for _, stat := range chessRes.Stats {
		if slices.Contains(keys, stat.Key) {
			message = strings.Replace(message, "="+stat.Key, string(stat.Stats.Rating), 1)
		}
	}

	logger.InfoContext(ctx, "request completed")
	fmt.Fprint(w, message)
}
