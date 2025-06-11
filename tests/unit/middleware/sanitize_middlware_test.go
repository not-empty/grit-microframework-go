package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestSanitizeMiddleware_RemovesQuotesFromQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?name=\"Jo'hn\"&city='New\"York'", nil)
	rr := httptest.NewRecorder()

	called := false
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		query := r.URL.Query()
		require.Equal(t, "John", query.Get("name"))
		require.Equal(t, "NewYork", query.Get("city"))
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestSanitizeMiddleware_RemovesQuotesFromHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User", "'Alice\" Smith'")
	req.Header.Add("X-Custom", `"test"`)

	rr := httptest.NewRecorder()

	called := false
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Equal(t, "Alice Smith", r.Header.Get("X-User"))
		require.Equal(t, "test", r.Header.Get("X-Custom"))
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, rr.Code)
}
