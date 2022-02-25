package utils

import (
	"encoding/json"
	"net/http"
	"net/mail"
)

func RespondWithJSON(w http.ResponseWriter, code int32, player interface{}) {
	response, _ := json.Marshal(player)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(code))
	w.Write(response)
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidResult(result float64) bool {
	return result > 0
}
