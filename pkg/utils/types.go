package utils

import "encoding/json"

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
