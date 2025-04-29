package helper

import (
	"encoding/json"
	"net/http"

	"github.com/not-empty/grit/app/config"
)

type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

func JSONErrorSimple(w http.ResponseWriter, status int, message string) {
	JSONError(w, status, message, nil)
}

func JSONError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	appEnv := config.AppConfig.AppEnv

	response := ErrorResponse{
		Error:  message,
		Detail: "",
	}

	if err != nil {
		if appEnv == "local" {
			response.Detail = err.Error()
		} else {
			response.Detail = ""
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}
