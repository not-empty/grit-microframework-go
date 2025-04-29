package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/not-empty/grit/app/config"
	"github.com/not-empty/grit/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestRecoverMiddleware_PanicProduction(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	_ = config.LoadConfig()

	req := httptest.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	handler := middleware.RecoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom!")
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var body map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)

	require.Equal(t, "Internal Server Error", body["error"])
	require.Nil(t, body["panic_error"])
	require.Nil(t, body["stacktrace"])
}

func TestRecoverMiddleware_PanicLocalIncludesStacktrace(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	_ = config.LoadConfig()

	req := httptest.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	handler := middleware.RecoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test crash")
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var body map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)

	require.Equal(t, "Internal Server Error", body["error"])
	require.Contains(t, body["panic_error"], "test crash")
	require.True(t, strings.Contains(body["stacktrace"].(string), "runtime/debug.Stack"))
}
