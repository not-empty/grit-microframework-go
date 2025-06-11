package middleware

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	appctx "github.com/not-empty/grit-microframework-go/app/context"

	"github.com/not-empty/grit-microframework-go/app/config"
)

type statusRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.StatusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.AppConfig.AppLog {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, StatusCode: 200}
		next.ServeHTTP(rec, r)

		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}

		queryDecoded, err := url.QueryUnescape(r.URL.RawQuery)
		if err != nil {
			queryDecoded = r.URL.RawQuery
		}

		fullPath := r.URL.Path
		if queryDecoded != "" {
			fullPath += "?" + queryDecoded
		}

		reqID, _ := r.Context().Value(appctx.RequestIDKey).(string)

		log.Printf("%s [%s] %s %s %d %s", reqID, r.Method, ip, fullPath, rec.StatusCode, time.Since(start))
	})
}
