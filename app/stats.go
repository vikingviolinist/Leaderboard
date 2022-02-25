package app

import (
	"fmt"
	"net/http"

	"github.com/kindly-ai/pingpong-leaderboard/models"
	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

func (a *App) getPlayerWins(w http.ResponseWriter, r *http.Request) {
	var status int32
	var player models.Player

	if status, player = models.NewPlayerFromBody(r.Body); status > 0 {
		utils.RespondWithJSON(w, status, nil)
		return
	}

	if !models.RedisKeyExists(a.Redis, key, player.Email) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(player.Email+"-wins", 0, -1).Val()

	players := []models.Player{}
	for _, data := range zRevRangeWithScores {
		member, _ := data.Member.(string)

		player := models.NewPlayer(member, data.Score)
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) getPlayerLosses(w http.ResponseWriter, r *http.Request) {
	var status int32
	var player models.Player

	if status, player = models.NewPlayerFromBody(r.Body); status > 0 {
		utils.RespondWithJSON(w, status, nil)
		return
	}

	if !models.RedisKeyExists(a.Redis, key, player.Email) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(player.Email+"-losses", 0, -1).Val()

	players := []models.Player{}
	for _, data := range zRevRangeWithScores {
		member, _ := data.Member.(string)

		player := models.Player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) updateScore(w http.ResponseWriter, r *http.Request) {
	var status int32
	var match models.Match

	if status, match = models.NewMatchFromBody(r.Body); status > 0 {
		utils.RespondWithJSON(w, status, nil)
		return
	}

	if !models.RedisKeyExists(a.Redis, key, match.Winner) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}
	if !models.RedisKeyExists(a.Redis, key, match.Loser) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	a.Redis.ZIncrBy(key, match.Result, match.Winner)
	a.Redis.ZIncrBy(match.Winner+"-wins", match.Result, match.Loser)
	a.Redis.ZIncrBy(match.Loser+"-losses", match.Result, match.Winner)

	fmt.Printf("Increased score for player %s by %.2f\n", match.Winner, match.Result)
	utils.RespondWithJSON(w, http.StatusOK, match)
}

func (a *App) getRank(w http.ResponseWriter, r *http.Request) {
	var player models.Player
	var status int32

	if status, player = models.NewPlayerFromBody(r.Body); status > 0 {
		utils.RespondWithJSON(w, status, nil)
		return
	}

	if !models.RedisKeyExists(a.Redis, key, player.Email) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	zRank := a.Redis.ZRevRank(key, player.Email).Val() + 1

	fmt.Printf("User %s is number %d\n", player.Email, zRank)
	utils.RespondWithJSON(w, http.StatusOK, zRank)
}

func (a *App) getTopThree(w http.ResponseWriter, r *http.Request) {
	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(key, 0, 2).Val()

	players := []models.Player{}
	for _, data := range zRevRangeWithScores {
		member, _ := data.Member.(string)

		player := models.NewPlayer(member, data.Score)
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}
