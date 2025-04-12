package middleware

import (
	"context"
	"net/http"

	"github.com/not-empty/grit/app/util/ulid"

	appctx "github.com/not-empty/grit/app/context"
)

func IdMiddleware(next http.Handler) http.Handler {
	var u ulid.Ulid

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := u.Generate(0)
		ctx := r.Context()
		ctx = context.WithValue(ctx, appctx.RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
