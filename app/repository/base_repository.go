package repository

import (
	"context"
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
	Add(ctx context.Context, m T) error
	Bulk(ctx context.Context, ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error)
	BulkAdd(ctx context.Context, models []T) error
	DeadDetail(ctx context.Context, id interface{}, fields []string) (map[string]any, error)
	DeadList(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	Delete(ctx context.Context, m T) error
	Detail(ctx context.Context, id interface{}, fields []string) (map[string]any, error)
	Edit(ctx context.Context, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error
	List(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	ListOne(ctx context.Context, orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error)
	Raw(ctx context.Context, query string, params map[string]any) ([]map[string]any, error)
	Undelete(ctx context.Context, m T) error
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

func (r *Repository[T]) Add(ctx context.Context, m T) error {
	return addRecord(ctx, r.DB, m)
}

func (r *Repository[T]) BulkAdd(ctx context.Context, m []T) error {
	baseModels := make([]BaseModel, len(m))
	for i, model := range m {
		baseModels[i] = model
	}
	return bulkAddRecords(ctx, r.DB, baseModels)
}

func (r *Repository[T]) Bulk(ctx context.Context, ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	m := r.New()
	return bulkRecords(ctx, r.DB, m.Schema(), m.TableName(), m.PrimaryKey(), fields, ids, limit, pageCursor, orderBy, order)
}

func (r *Repository[T]) DeadDetail(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(ctx, r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, true)
}

func (r *Repository[T]) DeadList(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(ctx, r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, true)
}

func (r *Repository[T]) Delete(ctx context.Context, m T) error {
	return deleteRecord(ctx, r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}

func (r *Repository[T]) Detail(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(ctx, r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, false)
}

func (r *Repository[T]) Edit(ctx context.Context, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	return editRecord(ctx, r.DB, table, pk, pkVal, cols, vals)
}

func (r *Repository[T]) List(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(ctx, r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, false)
}

func (r *Repository[T]) ListOne(ctx context.Context, orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error) {
	results, err := r.List(ctx, 1, nil, orderBy, order, fields, filters)
	if len(results) == 0 {
		return make(map[string]any), err
	}
	return results[0], err
}

func (r *Repository[T]) Raw(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	m := r.New()
	sqlText, args := helper.PrepareRawQuery(query, params)
	return rawRecords(ctx, r.DB, m.Schema(), sqlText, args...)
}

func (r *Repository[T]) Undelete(ctx context.Context, m T) error {
	return undeleteRecord(ctx, r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}
