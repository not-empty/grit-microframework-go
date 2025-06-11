package middleware

import (
	"net/http"

	"github.com/not-empty/grit-microframework-go/app/helper"
)

func ClosedChain(handler http.Handler) http.Handler {
	return helper.Chain(
		handler,
		LogMiddleware,
		RecoverMiddleware,
		IdMiddleware,
		AuthMiddleware,
		ResponseMiddleware,
		CorsMiddleware,
		SanitizeMiddleware,
	)
}

func AuthChain(handler http.Handler) http.Handler {
	return helper.Chain(
		handler,
		LogMiddleware,
		RecoverMiddleware,
		IdMiddleware,
		CorsMiddleware,
	)
}

func OpenChain(handler http.Handler) http.Handler {
	return helper.Chain(
		handler,
		LogMiddleware,
		RecoverMiddleware,
		IdMiddleware,
		ResponseMiddleware,
		CorsMiddleware,
		SanitizeMiddleware,
	)
}
