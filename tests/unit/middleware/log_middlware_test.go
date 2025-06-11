package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appctx "github.com/not-empty/grit-microframework-go/app/context"

	"github.com/not-empty/grit-microframework-go/app/config"
	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/stretchr/testify/require"
)

func TestLogMiddleware_SkipsLoggingWhenDisabled(t *testing.T) {
	os.Setenv("APP_LOG", "false")
	defer os.Unsetenv("APP_LOG")

	_ = config.LoadConfig()

	called := false
	req := httptest.NewRequest("GET", "/test-path", nil)
	rr := httptest.NewRecorder()

	handler := middleware.LogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))

	handler.ServeHTTP(rr, req)
	require.True(t, called)
	require.Equal(t, http.StatusTeapot, rr.Code)
}

func TestLogMiddleware_LogsOutputWhenEnabled(t *testing.T) {
	os.Setenv("APP_LOG", "true")
	defer os.Unsetenv("APP_LOG")

	_ = config.LoadConfig()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest("GET", "/hello?x=1&y=2", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	ctx := context.WithValue(req.Context(), appctx.RequestIDKey, "ulid-123")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := middleware.LogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))

	handler.ServeHTTP(rr, req)

	logLine := buf.String()
	require.Contains(t, logLine, "ulid-123")
	require.Contains(t, logLine, "[GET]")
	require.Contains(t, logLine, "127.0.0.1")
	require.Contains(t, logLine, "/hello?x=1&y=2")
	require.Contains(t, logLine, "202")
}

func TestLogMiddleware_QueryUnescapeFails(t *testing.T) {
	os.Setenv("APP_LOG", "true")
	defer os.Unsetenv("APP_LOG")

	_ = config.LoadConfig()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest("GET", "/broken?q=%zz", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req = req.WithContext(context.WithValue(req.Context(), appctx.RequestIDKey, "rid-1"))

	rr := httptest.NewRecorder()

	handler := middleware.LogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	require.Contains(t, buf.String(), "/broken?q=%zz")
}
