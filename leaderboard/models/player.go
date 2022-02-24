package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

type Player struct {
	Email string  `json:"email"`
	Score float64 `json:"score"`
}

func (p *Player) IsValid() bool {
	return utils.IsValidEmail(p.Email)
}

func NewPlayer(email string, score float64) *Player {
	return &Player{
		Email: email,
		Score: score,
	}
}

func NewPlayerFromBody(body io.Reader) (int32, *Player) {
	playerToCreate := NewPlayer("", 0.0)
	if resErr := json.NewDecoder(body).Decode(playerToCreate); resErr != nil {
		fmt.Printf("Invalid parameters\n")
		return http.StatusBadRequest, &Player{}
	}

	if !playerToCreate.IsValid() {
		fmt.Printf("Invalid parameters\n")
		return http.StatusBadRequest, &Player{}
	}

	return 0, playerToCreate
}
