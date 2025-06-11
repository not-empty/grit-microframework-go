package route

import (
	"net/http"

	"github.com/not-empty/grit-microframework-go/app/controller"
	"github.com/not-empty/grit-microframework-go/app/middleware"
	"github.com/not-empty/grit-microframework-go/app/router/registry"
)

func init() {
	ctrl := controller.NewAuthController("")
	registry.RegisterRoute("/auth/generate", middleware.AuthChain(http.HandlerFunc(ctrl.Generate)))
}
