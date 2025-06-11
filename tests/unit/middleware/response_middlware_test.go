package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/stretchr/testify/require"

	appctx "github.com/not-empty/grit-microframework-go/app/context"
)

func TestResponseMiddleware_InjectsHeadersAndBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/profile", nil)

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

func TestResponseMiddleware_SetsPageCursorHeader(t *testing.T) {
	const limit = helper.DefaultPageLimit
	arr := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		arr[i] = map[string]interface{}{
			"id":    fmt.Sprintf("%d", i+1),
			"value": fmt.Sprintf("value%d", i+1),
		}
	}
	bodyBytes, err := json.Marshal(arr)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	handler := middleware.ResponseMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}))

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	cursor := rr.Header().Get("X-Page-Cursor")
	require.NotEmpty(t, cursor, "expected X-Page-Cursor header")

	pc, err := helper.DecodeCursor(cursor)
	require.NoError(t, err)
	require.Equal(t, "25", pc.LastID)
	require.Equal(t, "25", pc.LastValue)

	req2 := httptest.NewRequest("GET", "/foo?order_by=value", bytes.NewReader(bodyBytes))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)

	cursor2 := rr2.Header().Get("X-Page-Cursor")
	require.NotEmpty(t, cursor2)

	pc2, err := helper.DecodeCursor(cursor2)
	require.NoError(t, err)
	require.Equal(t, "25", pc2.LastID)
	require.Equal(t, "value25", pc2.LastValue)
}
