package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	appctx "github.com/not-empty/grit/app/context"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rr *responseRecorder) Header() http.Header {
	return rr.ResponseWriter.Header()
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.status == 0 {
		rr.status = http.StatusOK
	}
	return rr.body.Write(b)
}

func ResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rr, r)

		elapsed := time.Since(start).Seconds()

		jwtInfo := appctx.JwtTokenInfo{}
		if info := r.Context().Value(appctx.JwtContextKey); info != nil {
			if tokenInfo, ok := info.(appctx.JwtTokenInfo); ok {
				jwtInfo = tokenInfo
			}
		}

		requestID := ""
		if rid := r.Context().Value(appctx.RequestIDKey); rid != nil {
			if str, ok := rid.(string); ok {
				requestID = str
			}
		}

		// Set headers
		w.Header().Set("X-Token", jwtInfo.Token)
		w.Header().Set("X-Expires", jwtInfo.Expires)
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Profile", formatProfile(elapsed))

		// Handle 204 manually
		if rr.status == http.StatusNoContent {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rr.status)
		w.Write(rr.body.Bytes())
	})
}

func formatProfile(sec float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.9f", sec), "0"), ".")
}
