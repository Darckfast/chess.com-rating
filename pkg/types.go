package chess

import "encoding/json"

type ChessStats struct {
	Stats []struct {
		Key   string `json:"key"`
		Stats struct {
			Rating json.RawMessage `json:"rating"`
		} `json:"stats"`
	} `json:"stats"`
}
