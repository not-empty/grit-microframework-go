package middleware

import (
	"net/http"
	"strings"
)

func sanitizeInputValues(values []string) []string {
	cleaned := make([]string, len(values))
	for i, v := range values {
		clean := strings.ReplaceAll(v, `"`, "")
		clean = strings.ReplaceAll(clean, `'`, "")
		cleaned[i] = clean
	}
	return cleaned
}

func SanitizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for key, values := range query {
			query[key] = sanitizeInputValues(values)
		}
		r.URL.RawQuery = query.Encode()

		for key, values := range r.Header {
			r.Header[key] = sanitizeInputValues(values)
		}

		next.ServeHTTP(w, r)
	})
}
