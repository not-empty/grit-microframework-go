package route

import (
	"net/http"

	"github.com/not-empty/grit/app/controller"
	"github.com/not-empty/grit/app/middleware"
	"github.com/not-empty/grit/app/router/registry"
)

func init() {
	ctrl := controller.NewAuthController("")
	registry.RegisterRoute("/auth/generate", middleware.AuthChain(http.HandlerFunc(ctrl.Generate)))
}
