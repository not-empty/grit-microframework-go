package middleware

import (
	"encoding/json"
	"net/http"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessMethods := "POST, GET, OPTIONS, PUT, DELETE, PATCH"
		accessHeaders := "Content-Type, Accept, Accept-Language, Authorization, X-Requested-With, Context, Suffix"

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", accessMethods)
		w.Header().Set("Access-Control-Allow-Headers", accessHeaders)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")

		if r.Method == http.MethodOptions {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"method": "OPTIONS",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
