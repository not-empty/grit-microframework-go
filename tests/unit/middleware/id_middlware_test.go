package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	appctx "github.com/not-empty/grit/app/context"
	"github.com/not-empty/grit/app/middleware"
)

type errorULIDGenerator struct{}

func (e *errorULIDGenerator) IsValidFormat(ulidStr string) bool {
	return true
}

func (e *errorULIDGenerator) GetTimeFromUlid(ulidStr string) (int64, error) {
	return 0, nil
}

func (e *errorULIDGenerator) GetDateFromUlid(ulidStr string) (string, error) {
	return "", nil
}

func (e *errorULIDGenerator) GetRandomnessFromString(ulidStr string) (string, error) {
	return "", nil
}

func (e *errorULIDGenerator) IsDuplicatedTime(t int64) bool {
	return false
}

func (e *errorULIDGenerator) HasIncrementLastRandChars(duplicateTime bool) bool {
	return false
}

func (e *errorULIDGenerator) Generate(t int64) (string, error) {
	return "", errors.New("ULID generation error")
}

func (e *errorULIDGenerator) DecodeTime(timePart string) (int64, error) {
	return 0, nil
}

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

func TestIdMiddleware_Fallback(t *testing.T) {
	fakeGen := &errorULIDGenerator{}

	req := httptest.NewRequest("GET", "/some-path", nil)
	rr := httptest.NewRecorder()

	handler := middleware.IdMiddlewareWithGenerator(fakeGen, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(appctx.RequestIDKey)
		require.Equal(t, "unknown", id, "Expected fallback request ID to be 'unknown' when Generate returns an error")
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)
}
