package registry

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit/app/router/registry"
	"github.com/stretchr/testify/require"
)

func TestRegisterRouteInitializerAndInitRoutes(t *testing.T) {
	called := false

	initFunc := func(db *sql.DB) {
		called = true
	}

	registry.RegisterRouteInitializer(initFunc)
	registry.InitRoutes(nil)

	require.True(t, called, "Expected route initializer to be called")
}

func TestRegisterAndGetRegisteredRoutes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	registry.RegisterRoute("/test", handler)

	routes := registry.GetRegisteredRoutes()
	require.Contains(t, routes, "/test")

	// Optional: Test that the registered handler actually works
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	routes["/test"].ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}
