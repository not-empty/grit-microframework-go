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
		repo := &repository.Repository[*models.User]{
			DB: db,
			New: func() *models.User {
				return new(models.User)
			},
		}
		baseRoutes := &route.BaseRoutes[*models.User]{
			Repo:   repo,
			Prefix: "/user",
			SetPK: func(m *models.User, id string) {
				m.ID = id
			},
		}
		baseRoutes.RegisterRoutes()
	})
}
