package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	appctx "github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/middleware"
)

func TestIdMiddleware_SetsRequestIdInContext(t *testing.T) {
	var capturedRequestID string

	req := httptest.NewRequest("GET", "/some-path", nil)
	rr := httptest.NewRecorder()

	handler := middleware.IdMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(appctx.RequestIDKey)
		require.NotNil(t, id, "Request ID should be set in context")
		require.IsType(t, "", id, "Request ID should be a string")
		capturedRequestID = id.(string)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotEmpty(t, capturedRequestID, "Request ID should not be empty")
	require.Len(t, capturedRequestID, 26, "ULID should be 26 characters long")
}
