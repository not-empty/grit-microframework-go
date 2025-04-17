package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/not-empty/grit/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ExemptPaths_BypassJwt(t *testing.T) {
	exemptPaths := []string{"/health", "/panic", "/auth/generate"}
	for _, path := range exemptPaths {
		t.Run(path, func(t *testing.T) {
			called := false
			req := httptest.NewRequest("GET", path, nil)
			rr := httptest.NewRecorder()

			handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))

			handler.ServeHTTP(rr, req)

			require.True(t, called, "handler should be called for exempt path")
			require.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestAuthMiddleware_NoAuthModeTrue_BypassJwt(t *testing.T) {
	os.Setenv("APP_NO_AUTH", "true")
	defer os.Unsetenv("APP_NO_AUTH")

	called := false
	req := httptest.NewRequest("GET", "/any-non-exempt", nil)
	rr := httptest.NewRecorder()

	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(rr, req)

	require.True(t, called, "handler should be called when APP_NO_AUTH is true")
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_WithJwtMiddleware(t *testing.T) {
	os.Setenv("APP_NO_AUTH", "false")
	defer os.Unsetenv("APP_NO_AUTH")

	var jwtCalled, handlerCalled bool

	original := middleware.JwtMiddlewareFunc
	middleware.JwtMiddlewareFunc = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCalled = true
			next.ServeHTTP(w, r)
		})
	}
	defer func() { middleware.JwtMiddlewareFunc = original }()

	req := httptest.NewRequest("GET", "/non-exempt", nil)
	rr := httptest.NewRecorder()

	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	handler.ServeHTTP(rr, req)

	require.True(t, jwtCalled, "JwtMiddlewareFunc should be invoked")
	require.True(t, handlerCalled, "handler should be called after JwtMiddlewareFunc")
	require.Equal(t, http.StatusOK, rr.Code)
}
