package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/repository"
	"github.com/not-empty/grit/app/util/ulid"
)

type BaseController[T repository.BaseModel] struct {
	Repo    repository.RepositoryInterface[T]
	Prefix  string
	SetPK   func(m T, id string)
	ULIDGen ulid.Generator
}

func NewBaseController[T repository.BaseModel](repo repository.RepositoryInterface[T], prefix string, setPK func(m T, id string)) *BaseController[T] {
	bc := &BaseController[T]{
		Repo:    repo,
		Prefix:  prefix,
		SetPK:   setPK,
		ULIDGen: ulid.NewDefaultGenerator(),
	}
	return bc
}

func (bc *BaseController[T]) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	m := bc.Repo.New()
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	helper.SanitizeModel(m)
	if err := helper.ValidatePayload(w, m); err != nil {
		return
	}

	id, err := bc.ULIDGen.Generate(0)
	if err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "ULID error", err)
		return
	}
	bc.SetPK(m, id)

	now := time.Now()
	if c, ok := any(m).(repository.Creatable); ok {
		c.SetCreatedAt(now)
	}
	if u, ok := any(m).(repository.Updatable); ok {
		u.SetUpdatedAt(now)
	}

	if err := bc.Repo.Insert(m); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "Insert error", err)
		return
	}

	helper.JSONResponse(w, http.StatusCreated, map[string]string{"id": id})
}

func (bc *BaseController[T]) Bulk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.IDs) == 0 {
		helper.JSONError(w, http.StatusBadRequest, "Invalid or empty Ids list", err)
		return
	}

	orderBy, order := helper.GetOrderParams(r, "id")
	limit, pageCursor, err := helper.GetPaginationParams(r)
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Invalid Page Cursor", err)
		return
	}
	fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())

	list, err := bc.Repo.BulkGet(input.IDs, limit, pageCursor, orderBy, order, fields)
	if err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "Bulk error", err)
		return
	}
	helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
}

func (bc *BaseController[T]) DeadDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/dead_detail/")
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Missing Id", err)
		return
	}

	fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
	m, err := bc.Repo.GetDeleted(id, fields)
	if err != nil {
		helper.JSONError(w, http.StatusNotFound, "Fields error", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, helper.FilterJSON(m, fields))
}

func (bc *BaseController[T]) DeadList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	orderBy, order := helper.GetOrderParams(r, "id")
	limit, pageCursor, err := helper.GetPaginationParams(r)
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Invalid Page Cursor", err)
		return
	}
	fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
	filters := helper.GetFilters(r, bc.Repo.New().Columns())

	list, err := bc.Repo.ListDeleted(limit, pageCursor, orderBy, order, fields, filters)
	if err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "List error", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
}

func (bc *BaseController[T]) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/delete/")
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Missing Id", err)
		return
	}

	m := bc.Repo.New()
	bc.SetPK(m, id)

	if err := bc.Repo.Delete(m); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "Delete error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (bc *BaseController[T]) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/detail/")
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Missing Id", err)
		return
	}

	fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
	m, err := bc.Repo.Get(id, fields)
	if err != nil {
		helper.JSONError(w, http.StatusNotFound, "Detail error", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, helper.FilterJSON(m, fields))
}

func (bc *BaseController[T]) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/edit/")
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Missing Id", err)
		return
	}

	var patchData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patchData); err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Invalid data", err)
		return
	}

	fetched, err := bc.Repo.Get(id, bc.Repo.New().Columns())
	if err != nil {
		helper.JSONError(w, http.StatusNotFound, "Not found", err)
		return
	}

	for key, value := range patchData {
		fetched[key] = value
	}

	helper.SanitizeModel(fetched)

	allCols := bc.Repo.New().Columns()
	var updateCols []string
	var updateVals []interface{}
	for _, col := range allCols {
		if _, exists := patchData[col]; exists {
			updateCols = append(updateCols, col)
			updateVals = append(updateVals, fetched[col])
		}
	}

	if _, ok := any(bc.Repo.New()).(repository.Updatable); ok {
		updateCols = append(updateCols, "updated_at")
		updateVals = append(updateVals, time.Now())
	}

	m := bc.Repo.New()
	bc.SetPK(m, id)

	if err := bc.Repo.UpdateFields(m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue(), updateCols, updateVals); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "Edit error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (bc *BaseController[T]) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helper.JSONErrorSimple(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	orderBy, order := helper.GetOrderParams(r, "id")
	limit, pageCursor, err := helper.GetPaginationParams(r)
	if err != nil {
		helper.JSONError(w, http.StatusBadRequest, "Invalid Page Cursor", err)
		return
	}
	fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
	filters := helper.GetFilters(r, bc.Repo.New().Columns())

	list, err := bc.Repo.ListActive(limit, pageCursor, orderBy, order, fields, filters)
	if err != nil {
		helper.JSONError(w, http.StatusInternalServerError, "List error", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
}
