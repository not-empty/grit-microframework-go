package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/not-empty/grit-microframework-go/app/config"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				response := map[string]interface{}{
					"error": "Internal Server Error",
				}
				if config.AppConfig.AppEnv == "local" {
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
