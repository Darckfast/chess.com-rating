package chess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"main/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
	"github.com/syumai/workers/cloudflare/fetch"
)

const (
	CHESS_CALLBACK_URL = "https://www.chess.com/callback/member/stats/"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:     r,
		AxiomApiKey: os.Getenv("AXIOM_API_KEY"),
		ServiceName: os.Getenv("VERCEL_GIT_REPO_SLUG"),
		RequestGen: func(args multilogger.SendLogsArgs) {
			args.MaxQueue <- 1
			args.Wg.Add(1)

			req, _ := fetch.NewRequest(args.Ctx, args.Method, args.Url, bytes.NewBuffer(*args.Body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", args.Bearer)

			client := fetch.NewClient()

			go func() {
				defer args.Wg.Done()
				client.Do(req, nil)
				<-args.MaxQueue
			}()
		},
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	logger.InfoContext(ctx, "processing request")

	username := r.URL.Query().Get("username")
	message := r.URL.Query().Get("message")
	message, _ = url.QueryUnescape(message)
	keys := r.URL.Query()["key"]

	username, _ = url.QueryUnescape(username)
	username = strings.TrimSpace(username)

	if username == "" {
		w.WriteHeader(400)
		logger.WarnContext(ctx, "username is required", "status", 400)
		fmt.Fprint(w, "username is required")
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(400)
		logger.WarnContext(ctx, "stats keys are required", "status", 400)
		fmt.Fprint(w, "stats keys are required")
		return
	}

	client := fetch.NewClient()
	req, _ := fetch.NewRequest(r.Context(), "GET", CHESS_CALLBACK_URL+url.QueryEscape(username), nil)

	res, err := client.Do(req, nil)
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

	logger.InfoContext(ctx, "request completed", "status", 200)
	fmt.Fprint(w, message)
}
