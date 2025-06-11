package middleware

import (
	"context"
	"net/http"

	"github.com/not-empty/ulid-go-lib"

	appctx "github.com/not-empty/grit/app/context"
)

func IdMiddlewareWithGenerator(gen ulid.Generator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID, err := gen.Generate(0)
		if err != nil {
			reqID = "unknown"
		}
		ctx := context.WithValue(r.Context(), appctx.RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IdMiddleware(next http.Handler) http.Handler {
	return IdMiddlewareWithGenerator(ulid.NewDefaultGenerator(), next)
}
