package app

import (
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	Redis  *redis.Client
}

func (a *App) Initialize() {
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

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/player", a.createPlayer).Methods("POST")
	a.Router.HandleFunc("/score", a.updateScore).Methods("PATCH")
	a.Router.HandleFunc("/wins", a.getPlayerWins).Methods("GET")
	a.Router.HandleFunc("/losses", a.getPlayerLosses).Methods("GET")
	a.Router.HandleFunc("/player", a.removePlayer).Methods("DELETE")
	a.Router.HandleFunc("/players", a.getPlayers).Methods("GET")
	a.Router.HandleFunc("/rank", a.getRank).Methods("GET")
	a.Router.HandleFunc("/top", a.getTopThree).Methods("GET")
}
