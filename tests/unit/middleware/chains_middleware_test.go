package middleware_test

import (
	"net/http"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareChains_ConstructProperly(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := middleware.ClosedChain(dummy)
	require.NotNil(t, handler)

	handler = middleware.OpenChain(dummy)
	require.NotNil(t, handler)

	handler = middleware.AuthChain(dummy)
	require.NotNil(t, handler)
}
