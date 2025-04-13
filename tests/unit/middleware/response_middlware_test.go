package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	appctx "github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/middleware"
)

func TestResponseMiddleware_InjectsHeadersAndBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/profile", nil)

	// Add context with JWT and RequestID
	req = req.WithContext(context.WithValue(req.Context(), appctx.JwtContextKey, appctx.JwtTokenInfo{
		Token:   "token-123",
		Expires: "2099-01-01",
	}))
	req = req.WithContext(context.WithValue(req.Context(), appctx.RequestIDKey, "req-123"))

	rr := httptest.NewRecorder()

	handler := middleware.ResponseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.JSONEq(t, `{"ok":"true"}`, rr.Body.String())
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	require.Equal(t, "token-123", rr.Header().Get("X-Token"))
	require.Equal(t, "2099-01-01", rr.Header().Get("X-Expires"))
	require.Equal(t, "req-123", rr.Header().Get("X-Request-ID"))
	require.NotEmpty(t, rr.Header().Get("X-Profile"))
	require.True(t, strings.HasPrefix(rr.Header().Get("X-Profile"), "0."))
}

func TestResponseMiddleware_NoContextFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/nocontent", nil)
	rr := httptest.NewRecorder()

	handler := middleware.ResponseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"foo": "bar"})
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.JSONEq(t, `{"foo":"bar"}`, rr.Body.String())
	require.Equal(t, "", rr.Header().Get("X-Token"))
	require.Equal(t, "", rr.Header().Get("X-Expires"))
	require.Equal(t, "", rr.Header().Get("X-Request-ID"))
	require.NotEmpty(t, rr.Header().Get("X-Profile"))
}

func TestResponseMiddleware_NoContentShortCircuit(t *testing.T) {
	req := httptest.NewRequest("GET", "/nocontent", nil)
	rr := httptest.NewRecorder()

	handler := middleware.ResponseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Empty(t, rr.Body.String())
	require.Equal(t, "", rr.Header().Get("Content-Type"), "should not write content-type on 204")
}

func TestResponseMiddleware_CallsHeaderMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/header-check", nil)
	rr := httptest.NewRecorder()

	handler := middleware.ResponseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(http.StatusAccepted)
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, "yes", rr.Header().Get("X-Test"))
	require.Equal(t, http.StatusAccepted, rr.Code)
}
