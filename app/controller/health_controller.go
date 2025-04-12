package controller

import (
	"encoding/json"
	"net/http"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (hc *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Ok"})
}

func (hc *HealthController) Panic(w http.ResponseWriter, r *http.Request) {
	panic("This is a test panic")
}
