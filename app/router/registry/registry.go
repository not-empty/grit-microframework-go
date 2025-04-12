package registry

import (
	"database/sql"
	"net/http"
)

type RouteInitializer func(db *sql.DB)

var initializers []RouteInitializer
var routeMap = make(map[string]http.Handler)

// RegisterRouteInitializer is used by base routes and domain modules.
func RegisterRouteInitializer(initFunc RouteInitializer) {
	initializers = append(initializers, initFunc)
}

// InitRoutes calls all registered initializers (usually called in main.go).
func InitRoutes(db *sql.DB) {
	for _, initFunc := range initializers {
		initFunc(db)
	}
}

// RegisterRoute allows direct route registration with pre-wrapped handlers.
func RegisterRoute(path string, handler http.Handler) {
	routeMap[path] = handler
}

// GetRegisteredRoutes returns the map of all directly registered routes.
func GetRegisteredRoutes() map[string]http.Handler {
	return routeMap
}
