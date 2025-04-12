package middleware

import (
	"net/http"
	"os"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		noAuthMode := os.Getenv("APP_NO_AUTH")
		if noAuthMode == "" {
			noAuthMode = "false"
		}

		exemptPaths := map[string]struct{}{
			"/health":        {},
			"/panic":         {},
			"/auth/generate": {},
		}

		if _, ok := exemptPaths[r.URL.Path]; ok || noAuthMode == "true" {
			next.ServeHTTP(w, r)
			return
		}

		JwtMiddleware(next).ServeHTTP(w, r)
	})
}
