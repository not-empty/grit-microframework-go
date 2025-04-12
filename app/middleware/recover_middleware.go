package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				appEnv := os.Getenv("APP_ENV")
				response := map[string]interface{}{
					"error": "Internal Server Error",
				}
				if appEnv == "local" {
					response["panic_error"] = fmt.Sprintf("%v", rec)
					response["stacktrace"] = string(debug.Stack())
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
