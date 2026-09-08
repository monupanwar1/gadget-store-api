package httpx

import (
	"encoding/json"
	"net/http"
)

type error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type envelope struct {
	Error error `json:"error"`
}

func ErrorResponse(w http.ResponseWriter, status int, message string, code string) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(envelope{
		Error: error{
			Code:    code,
			Message: message,
		},
	})

}
