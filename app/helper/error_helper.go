package helper

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// Accepts either error or string as the last param
func JSONError(w http.ResponseWriter, status int, msg interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var message string

	switch v := msg.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = "Unknown error"
	}

	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
