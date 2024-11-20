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
)

type MemberCallback struct {
	Stats []struct {
		Key   string `json:"key"`
		Stats struct {
			Rating                json.RawMessage `json:"rating"`
			HighestRating         int             `json:"highest_rating"`
			HighestRatingDate     string          `json:"highest_rating_date"`
			RatingTimeChangeDays  int             `json:"rating_time_change_days"`
			RatingTimeChangeValue int             `json:"rating_time_change_value"`
			TotalGameCount        int             `json:"total_game_count"`
			TotalWinCount         int             `json:"total_win_count"`
			TotalLossCount        int             `json:"total_loss_count"`
			TotalDrawCount        int             `json:"total_draw_count"`
			AvgOpponentRating     int             `json:"avg_opponent_rating"`
			TimeoutPercent        int             `json:"timeout_percent"`
			TimeoutDays           int             `json:"timeout_days"`
			TotalInProgressCount  int             `json:"total_in_progress_count"`
			AvgMoveTime           float64         `json:"avg_move_time"`
			LastDate              string          `json:"last_date"`
		} `json:"stats"`
		GameCount  int    `json:"gameCount"`
		LastPlayed bool   `json:"lastPlayed"`
		LastDate   string `json:"lastDate,omitempty"`
	} `json:"stats"`
	LastType string `json:"lastType"`
	Versus   struct {
		Total int `json:"total"`
	} `json:"versus"`
	RatingOnlyStats []string    `json:"ratingOnlyStats"`
	OfficialRating  interface{} `json:"officialRating"`
	LessonLevel     struct {
		Icon     string `json:"icon"`
		Name     string `json:"name"`
		Progress int    `json:"progress"`
	} `json:"lessonLevel"`
}

const (
	CHESS_CALLBACK_URL = "https://www.chess.com/callback/member/stats/"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func Handler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Incoming request")
	username := r.URL.Query().Get("username")
	message := r.URL.Query().Get("message")
	message, _ = url.QueryUnescape(message)
	keys := r.URL.Query()["key"]

	username, _ = url.QueryUnescape(username)
	username = strings.TrimSpace(username)

	if username == "" {
		w.WriteHeader(400)

		fmt.Fprint(w, "username is required")
		return
	}

	if len(keys) == 0 {
		w.WriteHeader(400)

		fmt.Fprint(w, "stats keys are required")
		return
	}

	req, _ := http.NewRequest("GET", CHESS_CALLBACK_URL+url.QueryEscape(username), nil)

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

	var chessRes MemberCallback
	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&chessRes)

	for _, stat := range chessRes.Stats {
		if slices.Contains(keys, stat.Key) {
			message = strings.Replace(message, "="+stat.Key, string(stat.Stats.Rating), 1)
		}
	}

	fmt.Fprint(w, message)
}
