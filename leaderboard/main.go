package main

import (
	"net/http"

	"github.com/kindly-ai/pingpong-leaderboard/app"
)

func main() {
	a := app.App{}
	a.Initialize()
	http.ListenAndServe(":8000", a.Router)
}
