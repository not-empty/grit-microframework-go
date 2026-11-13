package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/controller"
	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/stretchr/testify/require"
)

type contextMarkerKey struct{}

func TestBaseControllerPropagatesRequestContext(t *testing.T) {
	helper.RegisterRawQueries("fake", map[string]string{
		"probe": "SELECT id FROM fake",
	})

	tests := []struct {
		name      string
		method    string
		path      string
		body      string
		operation string
		handler   func(*controller.BaseController[*fakeModel], http.ResponseWriter, *http.Request)
	}{
		{"add", http.MethodPost, "/fake/add", `{"field":"value"}`, "Add", (*controller.BaseController[*fakeModel]).Add},
		{"bulk", http.MethodPost, "/fake/bulk", `{"ids":["1"]}`, "Bulk", (*controller.BaseController[*fakeModel]).Bulk},
		{"bulk add", http.MethodPost, "/fake/bulk_add", `[{"field":"value"}]`, "BulkAdd", (*controller.BaseController[*fakeModel]).BulkAdd},
		{"dead detail", http.MethodGet, "/fake/dead_detail/1", "", "DeadDetail", (*controller.BaseController[*fakeModel]).DeadDetail},
		{"dead list", http.MethodGet, "/fake/dead_list", "", "DeadList", (*controller.BaseController[*fakeModel]).DeadList},
		{"delete", http.MethodDelete, "/fake/delete/1", "", "Delete", (*controller.BaseController[*fakeModel]).Delete},
		{"detail", http.MethodGet, "/fake/detail/1", "", "Detail", (*controller.BaseController[*fakeModel]).Detail},
		{"edit", http.MethodPatch, "/fake/edit/1", `{"field":"new value"}`, "Edit", (*controller.BaseController[*fakeModel]).Edit},
		{"list", http.MethodGet, "/fake/list", "", "List", (*controller.BaseController[*fakeModel]).List},
		{"list one", http.MethodGet, "/fake/list_one", "", "ListOne", (*controller.BaseController[*fakeModel]).ListOne},
		{"raw", http.MethodPost, "/fake/select_raw", `{"query":"probe","params":{}}`, "Raw", (*controller.BaseController[*fakeModel]).Raw},
		{"undelete", http.MethodPatch, "/fake/undelete/1", "", "Undelete", (*controller.BaseController[*fakeModel]).Undelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				getResult: map[string]any{"id": "1", "field": "value"},
			}
			ctrl := controller.NewBaseController[*fakeModel](repo, "/fake", func(m *fakeModel, id string) {
				m.ID = id
			})

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			ctx, cancel := context.WithCancel(context.WithValue(req.Context(), contextMarkerKey{}, "request-context"))
			cancel()
			req = req.WithContext(ctx)

			tt.handler(ctrl, httptest.NewRecorder(), req)

			require.Equal(t, tt.operation, repo.lastOperation)
			require.NotNil(t, repo.lastContext)
			require.Equal(t, "request-context", repo.lastContext.Value(contextMarkerKey{}))
			require.ErrorIs(t, repo.lastContext.Err(), context.Canceled)
		})
	}
}
