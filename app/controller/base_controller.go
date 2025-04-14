package controller

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"

	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/repository"
	"github.com/not-empty/grit/app/util/ulid"
)

type BaseController[T repository.BaseModel] struct {
	Repo   *repository.Repository[T]
	Prefix string
	SetPK  func(m T, id string)
}

func (bc *BaseController[T]) Add(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodPost {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		m := bc.Repo.New()
		if err := json.NewDecoder(r.Body).Decode(m); err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		helper.SanitizeModel(m)
		if err := helper.ValidatePayload(w, m); err != nil {
			return nil
		}

		var u ulid.Ulid
		id := u.Generate(0)
		bc.SetPK(m, id)

		now := time.Now()
		if c, ok := any(m).(repository.Creatable); ok {
			c.SetCreatedAt(now)
		}
		if u, ok := any(m).(repository.Updatable); ok {
			u.SetUpdatedAt(now)
		}

		if err := bc.Repo.Insert(m); err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}

		helper.JSONResponse(w, http.StatusCreated, map[string]string{"id": id})
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) Bulk(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodPost {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		var input struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.IDs) == 0 {
			helper.JSONError(w, http.StatusBadRequest, "Invalid or empty ids list")
			return nil
		}

		orderBy, order := helper.GetOrderParams(r, "id")
		limit, offset := helper.GetPaginationParams(r)
		fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())

		list, err := bc.Repo.BulkGet(input.IDs, limit, offset, orderBy, order, fields)
		if err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}
		helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) DeadDetail(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodGet {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/dead_detail/")
		if err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
		m, err := bc.Repo.GetDeleted(id, fields)
		if err != nil {
			helper.JSONError(w, http.StatusNotFound, err)
			return nil
		}

		helper.JSONResponse(w, http.StatusOK, helper.FilterJSON(m, fields))
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) DeadList(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodGet {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		orderBy, order := helper.GetOrderParams(r, "id")
		limit, offset := helper.GetPaginationParams(r)
		fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
		filters := helper.GetFilters(r, bc.Repo.New().Columns())

		list, err := bc.Repo.ListDeleted(limit, offset, orderBy, order, fields, filters)

		if err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}

		helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) Delete(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodDelete {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/delete/")
		if err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		m := bc.Repo.New()
		bc.SetPK(m, id)

		if err := bc.Repo.Delete(m); err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) Detail(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodGet {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/detail/")
		fmt.Println(id)
		if err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
		m, err := bc.Repo.Get(id, fields)
		if err != nil {
			helper.JSONError(w, http.StatusNotFound, err)
			return nil
		}

		helper.JSONResponse(w, http.StatusOK, helper.FilterJSON(m, fields))
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) Edit(w http.ResponseWriter, r *http.Request) {
	
	if err := func() error {
		if r.Method != http.MethodPatch {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		id, err := helper.ExtractID(r.URL.Path, bc.Prefix+"/edit/")
		if err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		var patchData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&patchData); err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		m := bc.Repo.New()
		bc.SetPK(m, id)

		fetched, err := bc.Repo.Get(id, []string{"id"})
		if err != nil {
			helper.JSONError(w, http.StatusNotFound, err)
			return nil
		}

		jsonData, _ := json.Marshal(patchData)
		if err := json.Unmarshal(jsonData, &fetched); err != nil {
			helper.JSONError(w, http.StatusBadRequest, err)
			return nil
		}

		helper.SanitizeModel(fetched)

		allCols := m.Columns()
		fieldIndex := make(map[string]int)
		for i, col := range allCols {
			fieldIndex[col] = i
		}

		var updateCols []string
		var updateVals []interface{}
		allVals := m.Values()

		for col := range patchData {
			if i, ok := fieldIndex[col]; ok {
				updateCols = append(updateCols, col)
				updateVals = append(updateVals, allVals[i])
			}
		}

		if _, ok := any(m).(repository.Updatable); ok {
			updateCols = append(updateCols, "updated_at")
			updateVals = append(updateVals, time.Now())
		}

		if err := bc.Repo.UpdateFields(m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue(), updateCols, updateVals); err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}

func (bc *BaseController[T]) List(w http.ResponseWriter, r *http.Request) {
	if err := func() error {
		if r.Method != http.MethodGet {
			helper.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return nil
		}

		orderBy, order := helper.GetOrderParams(r, "id")
		limit, offset := helper.GetPaginationParams(r)
		fields := helper.GetFieldsParam(r, bc.Repo.New().Columns())
		filters := helper.GetFilters(r, bc.Repo.New().Columns())

		list, err := bc.Repo.ListActive(limit, offset, orderBy, order, fields, filters)
		if err != nil {
			helper.JSONError(w, http.StatusInternalServerError, err)
			return nil
		}

		helper.JSONResponse(w, http.StatusOK, helper.FilterList(list, fields))
		return nil
	}(); err != nil {
		helper.JSONError(w, http.StatusInternalServerError, err)
	}
}
