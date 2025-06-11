package middleware

import (
	"net/http"

	"github.com/not-empty/grit-microframework-go/app/config"
)

var JwtMiddlewareFunc = JwtMiddleware

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		noAuthMode := config.AppConfig.AppNoAuth

		exemptPaths := map[string]struct{}{
			"/health":        {},
			"/panic":         {},
			"/auth/generate": {},
		}

		if _, ok := exemptPaths[r.URL.Path]; ok || noAuthMode {
			next.ServeHTTP(w, r)
			return
		}

		JwtMiddlewareFunc(next).ServeHTTP(w, r)
	})
}
