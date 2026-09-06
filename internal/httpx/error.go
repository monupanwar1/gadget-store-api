package httpx

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Error Error `json:"error"`
}

func error(w http.ResponseWriter, status int, message string, code string) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(Envelope{
		Error: Error{
			Code:    code,
			Message: message,
		},
	})

}
