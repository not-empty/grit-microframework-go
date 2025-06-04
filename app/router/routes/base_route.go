package route

import (
	"net/http"

	"github.com/not-empty/grit/app/controller"
	"github.com/not-empty/grit/app/middleware"
	"github.com/not-empty/grit/app/repository"
)

type BaseRoutes[T repository.BaseModel] struct {
	Repo   *repository.Repository[T]
	Prefix string
	SetPK  func(m T, id string)
}

func (br *BaseRoutes[T]) RegisterRoutes() {
	ctrl := controller.NewBaseController(br.Repo, br.Prefix, br.SetPK)

	http.Handle(br.Prefix+"/add", middleware.ClosedChain(http.HandlerFunc(ctrl.Add)))
	http.Handle(br.Prefix+"/bulk", middleware.ClosedChain(http.HandlerFunc(ctrl.Bulk)))
	http.Handle(br.Prefix+"/bulk_add", middleware.ClosedChain(http.HandlerFunc(ctrl.BulkAdd)))
	http.Handle(br.Prefix+"/dead_detail/", middleware.ClosedChain(http.HandlerFunc(ctrl.DeadDetail)))
	http.Handle(br.Prefix+"/dead_list", middleware.ClosedChain(http.HandlerFunc(ctrl.DeadList)))
	http.Handle(br.Prefix+"/delete/", middleware.ClosedChain(http.HandlerFunc(ctrl.Delete)))
	http.Handle(br.Prefix+"/detail/", middleware.ClosedChain(http.HandlerFunc(ctrl.Detail)))
	http.Handle(br.Prefix+"/edit/", middleware.ClosedChain(http.HandlerFunc(ctrl.Edit)))
	http.Handle(br.Prefix+"/list", middleware.ClosedChain(http.HandlerFunc(ctrl.List)))
	http.Handle(br.Prefix+"/list_one", middleware.ClosedChain(http.HandlerFunc(ctrl.ListOne)))
	http.Handle(br.Prefix+"/select_raw", middleware.ClosedChain(http.HandlerFunc(ctrl.Raw)))
}
