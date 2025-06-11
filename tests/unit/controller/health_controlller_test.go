package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/controller"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	hc := controller.NewHealthController()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	hc.Health(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var responseBody map[string]string
	err := json.NewDecoder(res.Body).Decode(&responseBody)
	require.NoError(t, err)
	require.Equal(t, "Ok", responseBody["status"])
}

func TestPanic(t *testing.T) {
	hc := controller.NewHealthController()

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	var recovered interface{}
	func() {
		defer func() {
			recovered = recover()
		}()
		hc.Panic(rr, req)
	}()

	require.NotNil(t, recovered, "Expected panic in Panic handler")
	require.Equal(t, "This is a test panic", recovered, "Unexpected panic message")
}
