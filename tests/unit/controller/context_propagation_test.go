package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/controller"
	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/stretchr/testify/require"

	ulidmock "github.com/not-empty/ulid-go-lib/mock"
)

type ctxProbeKey string

const ctxProbe ctxProbeKey = "ctx-probe"

type ctxSpyRepository struct {
	*fakeRepository
	seen map[string]context.Context
}

func newCtxSpy() *ctxSpyRepository {
	return &ctxSpyRepository{
		fakeRepository: &fakeRepository{
			getResult:         map[string]any{"id": "1", "field": "oldValue"},
			getDeletedResult:  map[string]any{"id": "1", "field": "oldValue"},
			listActiveResult:  []map[string]any{{"id": "1", "field": "value"}},
			listDeletedResult: []map[string]any{{"id": "1", "field": "value"}},
			bulkGetResult:     []map[string]any{{"id": "1", "field": "value"}},
			listOneResult:     map[string]any{"id": "1", "field": "value"},
			rawResult:         []map[string]any{{"id": "1"}},
		},
		seen: map[string]context.Context{},
	}
}

func (s *ctxSpyRepository) AddContext(ctx context.Context, m *fakeModel) error {
	s.seen["AddContext"] = ctx
	return s.fakeRepository.AddContext(ctx, m)
}

func (s *ctxSpyRepository) BulkAddContext(ctx context.Context, m []*fakeModel) error {
	s.seen["BulkAddContext"] = ctx
	return s.fakeRepository.BulkAddContext(ctx, m)
}

func (s *ctxSpyRepository) BulkContext(ctx context.Context, ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	s.seen["BulkContext"] = ctx
	return s.fakeRepository.BulkContext(ctx, ids, limit, pageCursor, orderBy, order, fields)
}

func (s *ctxSpyRepository) DeadDetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	s.seen["DeadDetailContext"] = ctx
	return s.fakeRepository.DeadDetailContext(ctx, id, fields)
}

func (s *ctxSpyRepository) DeadListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	s.seen["DeadListContext"] = ctx
	return s.fakeRepository.DeadListContext(ctx, limit, pageCursor, orderBy, order, fields, filters)
}

func (s *ctxSpyRepository) DeleteContext(ctx context.Context, m *fakeModel) error {
	s.seen["DeleteContext"] = ctx
	return s.fakeRepository.DeleteContext(ctx, m)
}

func (s *ctxSpyRepository) DetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	s.seen["DetailContext"] = ctx
	return s.fakeRepository.DetailContext(ctx, id, fields)
}

func (s *ctxSpyRepository) EditContext(ctx context.Context, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	s.seen["EditContext"] = ctx
	return s.fakeRepository.EditContext(ctx, table, pk, pkVal, cols, vals)
}

func (s *ctxSpyRepository) ListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	s.seen["ListContext"] = ctx
	return s.fakeRepository.ListContext(ctx, limit, pageCursor, orderBy, order, fields, filters)
}

func (s *ctxSpyRepository) ListOneContext(ctx context.Context, orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error) {
	s.seen["ListOneContext"] = ctx
	return s.fakeRepository.ListOneContext(ctx, orderBy, order, fields, filters)
}

func (s *ctxSpyRepository) RawContext(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	s.seen["RawContext"] = ctx
	return s.fakeRepository.RawContext(ctx, query, params)
}

func (s *ctxSpyRepository) UndeleteContext(ctx context.Context, m *fakeModel) error {
	s.seen["UndeleteContext"] = ctx
	return s.fakeRepository.UndeleteContext(ctx, m)
}

func newCtxSpyController(spy *ctxSpyRepository) *controller.BaseController[*fakeModel] {
	return &controller.BaseController[*fakeModel]{
		Repo:   spy,
		Prefix: "/fake",
		SetPK:  func(m *fakeModel, id string) { m.ID = id },
		ULIDGen: &ulidmock.ULIDMock{
			GenerateFunc: func(ts int64) (string, error) {
				return "00000000000000000000000000000", nil
			},
		},
	}
}

func mustJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func TestBaseController_RequestContextReachesRepository(t *testing.T) {
	helper.RegisterRawQueries("fake", map[string]string{"probe": "SELECT id FROM fake"})

	cases := []struct {
		name   string
		method string
		expect string
		body   func(t *testing.T) *bytes.Buffer
		call   func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request)
	}{
		{
			name: "Add", method: http.MethodPost, expect: "AddContext",
			body: func(t *testing.T) *bytes.Buffer { return mustJSON(t, map[string]string{"field": "value"}) },
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Add(rr, r)
			},
		},
		{
			name: "BulkAdd", method: http.MethodPost, expect: "BulkAddContext",
			body: func(t *testing.T) *bytes.Buffer {
				return mustJSON(t, []map[string]string{{"field": "value"}})
			},
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.BulkAdd(rr, r)
			},
		},
		{
			name: "Bulk", method: http.MethodPost, expect: "BulkContext",
			body: func(t *testing.T) *bytes.Buffer {
				return mustJSON(t, map[string]any{"ids": []string{"1", "2"}})
			},
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Bulk(rr, r)
			},
		},
		{
			name: "DeadDetail", method: http.MethodGet, expect: "DeadDetailContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.DeadDetail(rr, r)
			},
		},
		{
			name: "DeadList", method: http.MethodGet, expect: "DeadListContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.DeadList(rr, r)
			},
		},
		{
			name: "Delete", method: http.MethodDelete, expect: "DeleteContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Delete(rr, r)
			},
		},
		{
			name: "Detail", method: http.MethodGet, expect: "DetailContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Detail(rr, r)
			},
		},
		{
			name: "Edit", method: http.MethodPatch, expect: "EditContext",
			body: func(t *testing.T) *bytes.Buffer { return mustJSON(t, map[string]any{"field": "newValue"}) },
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Edit(rr, r)
			},
		},
		{
			name: "List", method: http.MethodGet, expect: "ListContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.List(rr, r)
			},
		},
		{
			name: "ListOne", method: http.MethodGet, expect: "ListOneContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.ListOne(rr, r)
			},
		},
		{
			name: "Raw", method: http.MethodPost, expect: "RawContext",
			body: func(t *testing.T) *bytes.Buffer {
				return mustJSON(t, map[string]any{"query": "probe", "params": map[string]any{}})
			},
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Raw(rr, r)
			},
		},
		{
			name: "Undelete", method: http.MethodPatch, expect: "UndeleteContext",
			call: func(bc *controller.BaseController[*fakeModel], rr *httptest.ResponseRecorder, r *http.Request) {
				bc.Undelete(rr, r)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spy := newCtxSpy()
			bc := newCtxSpyController(spy)

			var req *http.Request
			if tc.body != nil {
				req = httptest.NewRequest(tc.method, "/fake/probe/1", tc.body(t))
			} else {
				req = httptest.NewRequest(tc.method, "/fake/probe/1", nil)
			}
			req = req.WithContext(context.WithValue(req.Context(), ctxProbe, "probe-value"))

			tc.call(bc, httptest.NewRecorder(), req)

			ctx, ok := spy.seen[tc.expect]
			require.True(t, ok, "%s should call %s so the request context reaches the database layer", tc.name, tc.expect)
			require.Equal(t, "probe-value", ctx.Value(ctxProbe), "request context values must survive until the repository")
			require.Nil(t, ctx.Done(), "context must not be cancelable, GetContextWithoutCancel strips cancelation")

			deadline, hasDeadline := ctx.Deadline()
			require.False(t, hasDeadline, "context must not carry a deadline, got %v", deadline)
		})
	}
}
