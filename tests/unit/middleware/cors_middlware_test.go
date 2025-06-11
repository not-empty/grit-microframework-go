package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestCorsMiddleware_HeadersApplied(t *testing.T) {
	req := httptest.NewRequest("GET", "/any", nil)
	rr := httptest.NewRecorder()

	handler := middleware.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	headers := rr.Header()

	require.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	require.Contains(t, headers.Get("Access-Control-Allow-Methods"), "POST")
	require.Contains(t, headers.Get("Access-Control-Allow-Headers"), "Authorization")
	require.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
	require.Equal(t, "86400", headers.Get("Access-Control-Max-Age"))

	require.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	require.Equal(t, "no-referrer", headers.Get("Referrer-Policy"))
	require.Contains(t, headers.Get("Strict-Transport-Security"), "max-age")
	require.Contains(t, headers.Get("Content-Security-Policy"), "default-src")
}

func TestCorsMiddleware_OptionsRequest(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/preflight", nil)
	rr := httptest.NewRecorder()

	handler := middleware.CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call next on OPTIONS")
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var data map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	require.NoError(t, err)
	require.Equal(t, "OPTIONS", data["method"])
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}
