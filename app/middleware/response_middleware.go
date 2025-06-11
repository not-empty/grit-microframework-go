package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/not-empty/grit-microframework-go/app/helper"

	appctx "github.com/not-empty/grit-microframework-go/app/context"
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

		h := w.Header()
		h.Set("X-Token", jwtInfo.Token)
		h.Set("X-Expires", jwtInfo.Expires)
		h.Set("X-Request-ID", requestID)
		h.Set("X-Profile", formatProfile(elapsed))

		if rr.status != http.StatusNoContent {
			if cursor, err := helper.BuildPageCursor(rr.body.Bytes(), r.URL.Query()); err == nil && cursor != "" {
				h.Set("X-Page-Cursor", cursor)
			}
		}

		if rr.status == http.StatusNoContent {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.Set("Content-Type", "application/json")
		w.WriteHeader(rr.status)
		w.Write(rr.body.Bytes())
	})
}

func formatProfile(sec float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.9f", sec), "0"), ".")
}
