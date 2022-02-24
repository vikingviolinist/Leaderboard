package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type app struct {
	Router *mux.Router
	Redis  *redis.Client
}

type player struct {
	Email string  `json:"email"`
	Score float64 `json:"score"`
}

type UpdatePlayerRequest struct {
	Winner string  `json:"winner"`
	Loser  string  `json:"loser"`
	Result float64 `json:"result"`
}

type getRankRequest struct {
	Email string `json:"email"`
}

const key = "leaderboard"

func respondWithJSON(w http.ResponseWriter, code int, player interface{}) {
	response, _ := json.Marshal(player)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *app) initialize() {
	a.Router = mux.NewRouter()
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if _, err := a.Redis.Ping().Result(); err != nil {
		panic(err)
	}

	a.initializeRoutes()
}

func (a *app) initializeRoutes() {
	a.Router.HandleFunc("/player", a.getPlayer).Methods("GET")
	a.Router.HandleFunc("/wins", a.getPlayerWins).Methods("GET")
	a.Router.HandleFunc("/losses", a.getPlayerLosses).Methods("GET")
	a.Router.HandleFunc("/player", a.createPlayer).Methods("POST")
	a.Router.HandleFunc("/score", a.updateScore).Methods("PATCH")
	a.Router.HandleFunc("/player", a.removePlayer).Methods("DELETE")
	a.Router.HandleFunc("/players", a.getPlayers).Methods("GET")
	a.Router.HandleFunc("/rank", a.getRank).Methods("GET")
	a.Router.HandleFunc("/top", a.getTopThree).Methods("GET")
}

func main() {
	a := app{}
	a.initialize()
	http.ListenAndServe(":8000", a.Router)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (a *app) getPlayer(w http.ResponseWriter, r *http.Request) {
	return
}

func (a *app) getPlayerWins(w http.ResponseWriter, r *http.Request) {
	playerToUpdate := new(player)

	if err := json.NewDecoder(r.Body).Decode(playerToUpdate); err != nil {
		fmt.Printf("Invalid parameters\n")
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !isValidEmail(playerToUpdate.Email) {
		fmt.Printf("Invalid email %s\n", playerToUpdate.Email)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToUpdate.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToUpdate.Email)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(playerToUpdate.Email+"-wins", 0, -1)
	fmt.Println(playerToUpdate.Email, zRevRangeWithScores.Val())
	players := []player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	respondWithJSON(w, http.StatusOK, players)
}

func (a *app) createPlayer(w http.ResponseWriter, r *http.Request) {
	playerToCreate := new(player)

	if resErr := json.NewDecoder(r.Body).Decode(playerToCreate); resErr != nil {
		fmt.Printf("Invalid parameters\n")
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !isValidEmail(playerToCreate.Email) {
		fmt.Printf("Invalid email %s\n", playerToCreate.Email)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToCreate.Email)

	if _, scoreErr := zScore.Result(); scoreErr == nil {
		fmt.Printf("Player %s already exists\n", playerToCreate.Email)
		respondWithJSON(w, http.StatusOK, nil)
		return
	}

	a.Redis.ZAddNX(key, redis.Z{Member: playerToCreate.Email})

	fmt.Printf("Created player %s\n", playerToCreate.Email)

	respondWithJSON(w, http.StatusCreated, playerToCreate)
}

func (a *app) getPlayerLosses(w http.ResponseWriter, r *http.Request) {
	playerToUpdate := new(player)

	if resErr := json.NewDecoder(r.Body).Decode(playerToUpdate); resErr != nil {
		fmt.Printf("Invalid parameters\n")
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !isValidEmail(playerToUpdate.Email) {
		fmt.Printf("Invalid email %s\n", playerToUpdate.Email)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToUpdate.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToUpdate.Email)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(playerToUpdate.Email+"-losses", 0, -1)

	players := []player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	respondWithJSON(w, http.StatusOK, players)
}

func (a *app) updateScore(w http.ResponseWriter, r *http.Request) {
	playerToUpdate := new(UpdatePlayerRequest)

	if err := json.NewDecoder(r.Body).Decode(playerToUpdate); err != nil {
		fmt.Printf("Invalid parameters\n")
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if !isValidEmail(playerToUpdate.Winner) || !isValidEmail(playerToUpdate.Loser) {
		fmt.Printf("Invalid email(s): %s, %s\n", playerToUpdate.Winner, playerToUpdate.Loser)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	if playerToUpdate.Result < float64(1) {
		fmt.Printf("Invalid result value %f\n", playerToUpdate.Result)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScoreWinner := a.Redis.ZScore(key, playerToUpdate.Winner)
	zScoreLoser := a.Redis.ZScore(key, playerToUpdate.Loser)

	if _, err := zScoreWinner.Result(); err != nil {
		fmt.Printf("Player doesn't exist %s\n", playerToUpdate.Winner)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	if _, err := zScoreLoser.Result(); err != nil {
		fmt.Printf("Player doesn't exist %s\n", playerToUpdate.Loser)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	a.Redis.ZIncrBy(key, playerToUpdate.Result, playerToUpdate.Winner)
	a.Redis.ZIncrBy(playerToUpdate.Winner+"-wins", playerToUpdate.Result, playerToUpdate.Loser)
	a.Redis.ZIncrBy(playerToUpdate.Loser+"-losses", playerToUpdate.Result, playerToUpdate.Winner)

	fmt.Printf("Increased score for player %s by %.2f\n", playerToUpdate.Winner, playerToUpdate.Result)
	respondWithJSON(w, http.StatusOK, playerToUpdate)
}

func (a *app) removePlayer(w http.ResponseWriter, r *http.Request) {
	playerToRemove := new(player)
	json.NewDecoder(r.Body).Decode(playerToRemove)

	if !isValidEmail(playerToRemove.Email) {
		fmt.Printf("Invalid email %s\n", playerToRemove.Email)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, playerToRemove.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", playerToRemove.Email)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	a.Redis.ZRem(key, playerToRemove.Email)
	a.Redis.Del(playerToRemove.Email + "-wins")
	a.Redis.Del(playerToRemove.Email + "-losses")

	fmt.Printf("Removed player %s\n", playerToRemove.Email)
	respondWithJSON(w, http.StatusNoContent, playerToRemove)
}

func (a *app) getRank(w http.ResponseWriter, r *http.Request) {
	rank := new(getRankRequest)
	json.NewDecoder(r.Body).Decode(rank)

	if !isValidEmail(rank.Email) {
		fmt.Printf("Invalid email %s\n", rank.Email)
		respondWithJSON(w, http.StatusBadRequest, nil)
		return
	}

	zScore := a.Redis.ZScore(key, rank.Email)

	if _, err := zScore.Result(); err != nil {
		fmt.Printf("User %s doesn't exist\n", rank.Email)
		respondWithJSON(w, http.StatusNotFound, nil)
		return
	}

	zRank := a.Redis.ZRevRank(key, rank.Email)
	fmt.Printf("User %s is number %d\n", rank.Email, zRank.Val()+1)

	respondWithJSON(w, http.StatusOK, zRank.Val()+1)
}

func (a *app) getTopThree(w http.ResponseWriter, r *http.Request) {
	zRevRangeWithScores := a.Redis.ZRevRangeWithScores(key, 0, 2)

	players := []player{}
	for _, data := range zRevRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	respondWithJSON(w, http.StatusOK, players)
}

func (a *app) getPlayers(w http.ResponseWriter, r *http.Request) {
	zRangeWithScores := a.Redis.ZRangeWithScores(key, 0, -1)

	players := []player{}
	for _, data := range zRangeWithScores.Val() {
		member, _ := data.Member.(string)

		player := player{Email: member, Score: data.Score}
		players = append(players, player)
	}
	respondWithJSON(w, http.StatusOK, players)
}
