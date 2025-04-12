package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func JSONResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func ExtractID(path, prefix string) (string, error) {
	id := strings.TrimPrefix(path, prefix)
	if id == "" {
		return "", errors.New("Missing ID")
	}
	return id, nil
}

func FilterList[T any](list []T, fields []string) []map[string]interface{} {
	filtered := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		filtered = append(filtered, FilterJSON(item, fields))
	}
	return filtered
}

func SanitizeModel(m any) {
	if s, ok := m.(interface{ Sanitize() }); ok {
		s.Sanitize()
	}
}
