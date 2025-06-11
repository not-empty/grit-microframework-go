package router

import (
	"database/sql"
	"net/http"

	"github.com/not-empty/grit-microframework-go/app/router/registry"
)

func RegisterRoutes(db *sql.DB) {
	registry.InitRoutes(db)

	for path, handler := range registry.GetRegisteredRoutes() {
		http.Handle(path, handler)
	}
}
