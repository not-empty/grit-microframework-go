package domains

import (
	"database/sql"

	"github.com/not-empty/grit/app/repository"
	"github.com/not-empty/grit/app/repository/models"
	"github.com/not-empty/grit/app/router/routes"
	"github.com/not-empty/grit/app/router/registry"
)

func init() {
	registry.RegisterRouteInitializer(func(db *sql.DB) {
		repo := repository.NewRepository[*models.Example](db, func() *models.Example {
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
