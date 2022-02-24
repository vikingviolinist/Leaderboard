package utils

import (
	"encoding/json"
	"net/http"
	"net/mail"
)

func RespondWithJSON(w http.ResponseWriter, code int, player interface{}) {
	response, _ := json.Marshal(player)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
