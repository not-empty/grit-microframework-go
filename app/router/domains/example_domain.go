package domains

import (
	"database/sql"

	"github.com/not-empty/grit-microframework-go/app/repository"
	"github.com/not-empty/grit-microframework-go/app/repository/models"
	"github.com/not-empty/grit-microframework-go/app/router/registry"
	route "github.com/not-empty/grit-microframework-go/app/router/routes"
)

func init() {
	registry.RegisterRouteInitializer(func(db *sql.DB) {
		repo := repository.NewRepository(db, func() *models.Example {
			return new(models.Example)
		})
		baseRoutes := &route.BaseRoutes[*models.Example]{
			Repo:   repo,
			Prefix: "/example",
			SetPK: func(m *models.Example, id string) {
				m.ID = id
			},
		}
		baseRoutes.RegisterRoutes()
	})
}
