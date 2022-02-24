package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kindly-ai/pingpong-leaderboard/models"
	"github.com/kindly-ai/pingpong-leaderboard/utils"
)

type getRankRequest struct {
	Email string `json:"email"`
}

func (a *App) getPlayerWins(w http.ResponseWriter, r *http.Request) {
	var status int32
	var playerToUpdate *models.Player
	if status, playerToUpdate = models.NewPlayerFromBody(r.Body); status > 0 {
		fmt.Printf("Invalid parameters\n")
		utils.RespondWithJSON(w, status, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToUpdate.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToUpdate.Email)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(playerToUpdate.Email+"-wins", 0, -1)
	fmt.Println(playerToUpdate.Email, zRevRangeWithScores.Val())
	players := []*models.Player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := models.NewPlayer(member, data.Score)
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) getPlayerLosses(w http.ResponseWriter, r *http.Request) {
	var status int32
	var playerToUpdate *models.Player
	if status, playerToUpdate = models.NewPlayerFromBody(r.Body); status > 0 {
		fmt.Printf("Invalid parameters\n")
		utils.RespondWithJSON(w, status, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToUpdate.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToUpdate.Email)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(playerToUpdate.Email+"-losses", 0, -1)

	players := []models.Player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := models.Player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}

func (a *App) updateScore(w http.ResponseWriter, r *http.Request) {
	playerToUpdate := new(UpdatePlayerRequest)

	if err := json.NewDecoder(r.Body).Decode(playerToUpdate); err != nil {
		fmt.Printf("Invalid parameters\n")
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !utils.IsValidEmail(playerToUpdate.Winner) || !utils.IsValidEmail(playerToUpdate.Loser) {
		fmt.Printf("Invalid email(s): %s, %s\n", playerToUpdate.Winner, playerToUpdate.Loser)
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if playerToUpdate.Result < float64(1) {
		fmt.Printf("Invalid result value %f\n", playerToUpdate.Result)
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScoreWinner := a.Redis.ZScore(key, playerToUpdate.Winner)
	zScoreLoser := a.Redis.ZScore(key, playerToUpdate.Loser)

	if _, err := zScoreWinner.Result(); err != nil {
		fmt.Printf("Player doesn't exist %s\n", playerToUpdate.Winner)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	if _, err := zScoreLoser.Result(); err != nil {
		fmt.Printf("Player doesn't exist %s\n", playerToUpdate.Loser)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	a.Redis.ZIncrBy(key, playerToUpdate.Result, playerToUpdate.Winner)
	a.Redis.ZIncrBy(playerToUpdate.Winner+"-wins", playerToUpdate.Result, playerToUpdate.Loser)
	a.Redis.ZIncrBy(playerToUpdate.Loser+"-losses", playerToUpdate.Result, playerToUpdate.Winner)

	fmt.Printf("Increased score for player %s by %.2f\n", playerToUpdate.Winner, playerToUpdate.Result)
	utils.RespondWithJSON(w, http.StatusOK, playerToUpdate)
}

func (a *App) getRank(w http.ResponseWriter, r *http.Request) {
	rank := new(getRankRequest)
	json.NewDecoder(r.Body).Decode(rank)

	if !utils.IsValidEmail(rank.Email) {
		fmt.Printf("Invalid email %s\n", rank.Email)
		utils.RespondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, rank.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", rank.Email)
		utils.RespondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRank := a.Redis.ZRevRank(key, rank.Email)
	fmt.Printf("User %s is number %d\n", rank.Email, zRank.Val()+1)

	utils.RespondWithJSON(w, http.StatusOK, zRank.Val()+1)
}

func (a *App) getTopThree(w http.ResponseWriter, r *http.Request) {
	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(key, 0, 2)

	players := []*models.Player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := models.NewPlayer(member, data.Score)
		players = append(players, player)
	}
	utils.RespondWithJSON(w, http.StatusOK, players)
}
