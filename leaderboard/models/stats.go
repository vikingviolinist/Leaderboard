package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

type Match struct {
	Winner string  `json:"winner"`
	Loser  string  `json:"loser"`
	Result float64 `json:"result"`
}

func (m *Match) IsValid() bool {
	return utils.IsValidResult(m.Result)
}

func NewMatch(winner string, loser string, result float64) Match {
	return Match{
		Winner: winner,
		Loser:  loser,
		Result: result,
	}
}

func NewMatchFromBody(body io.Reader) (int32, Match) {
	matchToCreate := NewMatch("", "", 0.0)
	if resErr := json.NewDecoder(body).Decode(&matchToCreate); resErr != nil {
		fmt.Printf("Invalid parameters\n")
		return http.StatusBadRequest, Match{}
	}

	return 0, matchToCreate
}
