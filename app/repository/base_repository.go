package repository

import (
	"database/sql"
	"time"

	"github.com/not-empty/grit-microframework-go/app/helper"
)

type Sanitizable interface {
	Sanitize()
}

type Creatable interface {
	SetCreatedAt(time.Time)
}

type Updatable interface {
	SetUpdatedAt(time.Time)
}

type Scanner interface {
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
}

type BaseModel interface {
	TableName() string
	Columns() []string
	Values() []interface{}
	HasDefaultValue() []string
	PrimaryKey() string
	PrimaryKeyValue() interface{}
	Schema() map[string]string
}

type RepositoryInterface[T BaseModel] interface {
	New() T
	Add(m T) error
	Bulk(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error)
	BulkAdd(models []T) error
	DeadDetail(id interface{}, fields []string) (map[string]any, error)
	DeadList(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	Delete(m T) error
	Detail(id interface{}, fields []string) (map[string]any, error)
	Edit(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error
	List(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	ListOne(orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error)
	Raw(query string, params map[string]any) ([]map[string]any, error)
}

type Repository[T BaseModel] struct {
	DB      *sql.DB
	newFunc func() T
}

func NewRepository[T BaseModel](db *sql.DB, newFunc func() T) *Repository[T] {
	return &Repository[T]{
		DB:      db,
		newFunc: newFunc,
	}
}

func (r *Repository[T]) New() T {
	return r.newFunc()
}

func (r *Repository[T]) Add(m T) error {
	return addRecord(r.DB, m)
}

func (r *Repository[T]) BulkAdd(m []T) error {
	baseModels := make([]BaseModel, len(m))
	for i, model := range m {
		baseModels[i] = model
	}
	return bulkAddRecords(r.DB, baseModels)
}

func (r *Repository[T]) Bulk(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	m := r.New()
	return bulkRecords(r.DB, m.Schema(), m.TableName(), m.PrimaryKey(), fields, ids, limit, pageCursor, orderBy, order)
}

func (r *Repository[T]) DeadDetail(id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, true)
}

func (r *Repository[T]) DeadList(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, true)
}

func (r *Repository[T]) Delete(m T) error {
	return deleteRecord(r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}

func (r *Repository[T]) Detail(id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, false)
}

func (r *Repository[T]) Edit(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	return editRecord(r.DB, table, pk, pkVal, cols, vals)
}

func (r *Repository[T]) List(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, false)
}

func (r *Repository[T]) ListOne(orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error) {
	results, err := r.List(1, nil, orderBy, order, fields, filters)
	if len(results) == 0 {
		return make(map[string]any), err
	}
	return results[0], err
}

func (r *Repository[T]) Raw(query string, params map[string]any) ([]map[string]any, error) {
	m := r.New()
	sqlText, args := helper.PrepareRawQuery(query, params)
	return rawRecords(r.DB, m.Schema(), sqlText, args...)
}
