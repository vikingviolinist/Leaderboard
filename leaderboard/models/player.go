package models

import "github.com/kindly-ai/pingpong-leaderboard/utils"

type Player struct {
	Email string  `json:"email"`
	Score float64 `json:"score"`
}

func (p *Player) IsValid() bool {
	return utils.IsValidEmail(p.Email)
}
