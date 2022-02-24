package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/kindly-ai/pingpong-leaderboard/models"
	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

const key = "leaderboard"

func (a *App) createPlayer(w http.ResponseWriter, r *http.Request) {
	playerToCreate := new(models.Player)

	if resErr := json.NewDecoder(r.Body).Decode(playerToCreate); resErr != nil {
		fmt.Printf("Invalid parameters\n")
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !utils.IsValidEmail(playerToCreate.Email) {
		fmt.Printf("Invalid email %s\n", playerToCreate.Email)
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToCreate.Email)

	if _, scoreErr := zScore.Result(); scoreErr == nil {
		fmt.Printf("Player %s already exists\n", playerToCreate.Email)
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	a.Redis.ZAddNX(key, redis.Z{Member: playerToCreate.Email})

	fmt.Printf("Created player %s\n", playerToCreate.Email)

	utils.RespondWithJSON(w, http.StatusCreated, playerToCreate)
}

func (a *App) getPlayers(w http.ResponseWriter, r *http.Request) {
	zRangeWithScores := a.Redis.ZRangeWithScores(key, 0, -1)

	players := []models.Player{}
	for _, data := range zRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := models.Player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) removePlayer(w http.ResponseWriter, r *http.Request) {
	playerToRemove := new(models.Player)
	json.NewDecoder(r.Body).Decode(playerToRemove)

	if !utils.IsValidEmail(playerToRemove.Email) {
		fmt.Printf("Invalid email %s\n", playerToRemove.Email)
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToRemove.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToRemove.Email)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	a.Redis.ZRem(key, playerToRemove.Email)
	a.Redis.Del(playerToRemove.Email + "-wins")
	a.Redis.Del(playerToRemove.Email + "-losses")

	fmt.Printf("Removed player %s\n", playerToRemove.Email)
	utils.RespondWithJSON(w, http.StatusNoContent, playerToRemove)
}
