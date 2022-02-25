package app

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/kindly-ai/pingpong-leaderboard/models"
	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

const key = "leaderboard"

func (a *App) createPlayer(w http.ResponseWriter, r *http.Request) {
	var player models.Player
	var status int32

	if status, player = models.NewPlayerFromBody(r.Body); status > 0 {
		utils.RespondWithJSON(w, status, nil)
		return
	}

	if models.RedisKeyExists(a.Redis, key, player.Email) {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	if err := models.RedisKeyCreate(a.Redis, key, redis.Z{Member: player.Email}); err != nil {
		utils.RespondWithJSON(w, http.StatusOK, player)
		return
	}

	fmt.Printf("Created player %s\n", player.Email)

	utils.RespondWithJSON(w, http.StatusCreated, player)
}

func (a *App) getPlayers(w http.ResponseWriter, r *http.Request) {
	zRangeWithScores := a.Redis.ZRangeWithScores(key, 0, -1)

	players := []models.Player{}
	for _, data := range zRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := models.NewPlayer(member, data.Score)
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) removePlayer(w http.ResponseWriter, r *http.Request) {
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

	a.Redis.ZRem(key, player.Email)
	a.Redis.Del(player.Email + "-wins")
	a.Redis.Del(player.Email + "-losses")

	fmt.Printf("Removed player %s\n", player.Email)
	utils.RespondWithJSON(w, http.StatusNoContent, player)
}
