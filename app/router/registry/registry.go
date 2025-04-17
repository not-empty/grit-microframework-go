package registry

import (
	"database/sql"
	"net/http"
)

type RouteInitializer func(db *sql.DB)

var initializers []RouteInitializer
var routeMap = make(map[string]http.Handler)

func RegisterRouteInitializer(initFunc RouteInitializer) {
	initializers = append(initializers, initFunc)
}

func InitRoutes(db *sql.DB) {
	for _, initFunc := range initializers {
		initFunc(db)
	}
}

func RegisterRoute(path string, handler http.Handler) {
	routeMap[path] = handler
}

func GetRegisteredRoutes() map[string]http.Handler {
	return routeMap
}
