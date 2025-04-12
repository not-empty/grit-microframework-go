package route

import (
	"net/http"

	"github.com/not-empty/grit/app/controller"
	"github.com/not-empty/grit/app/middleware"
	"github.com/not-empty/grit/app/router/registry"
)

func init() {
	ctrl := controller.NewHealthController()
	registry.RegisterRoute("/health", middleware.OpenChain(http.HandlerFunc(ctrl.Health)))
	registry.RegisterRoute("/panic", middleware.OpenChain(http.HandlerFunc(ctrl.Panic)))
}
