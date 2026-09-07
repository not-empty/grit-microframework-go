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
	Undelete(m T) error
	AddContext(ctx context.Context, m T) error
	BulkContext(ctx context.Context, ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error)
	BulkAddContext(ctx context.Context, models []T) error
	DeadDetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error)
	DeadListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	DeleteContext(ctx context.Context, m T) error
	DetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error)
	EditContext(ctx context.Context, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error
	ListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	ListOneContext(ctx context.Context, orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error)
	RawContext(ctx context.Context, query string, params map[string]any) ([]map[string]any, error)
	UndeleteContext(ctx context.Context, m T) error
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
	return r.AddContext(context.Background(), m)
}

func (r *Repository[T]) BulkAdd(m []T) error {
	return r.BulkAddContext(context.Background(), m)
}

func (r *Repository[T]) Bulk(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	return r.BulkContext(context.Background(), ids, limit, pageCursor, orderBy, order, fields)
}

func (r *Repository[T]) DeadDetail(id interface{}, fields []string) (map[string]any, error) {
	return r.DeadDetailContext(context.Background(), id, fields)
}

func (r *Repository[T]) DeadList(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	return r.DeadListContext(context.Background(), limit, pageCursor, orderBy, order, fields, filters)
}

func (r *Repository[T]) Delete(m T) error {
	return r.DeleteContext(context.Background(), m)
}

func (r *Repository[T]) Detail(id interface{}, fields []string) (map[string]any, error) {
	return r.DetailContext(context.Background(), id, fields)
}

func (r *Repository[T]) Edit(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	return r.EditContext(context.Background(), table, pk, pkVal, cols, vals)
}

func (r *Repository[T]) List(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	return r.ListContext(context.Background(), limit, pageCursor, orderBy, order, fields, filters)
}

func (r *Repository[T]) ListOne(orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error) {
	return r.ListOneContext(context.Background(), orderBy, order, fields, filters)
}

func (r *Repository[T]) Raw(query string, params map[string]any) ([]map[string]any, error) {
	return r.RawContext(context.Background(), query, params)
}

func (r *Repository[T]) Undelete(m T) error {
	return r.UndeleteContext(context.Background(), m)
}

func (r *Repository[T]) AddContext(ctx context.Context, m T) error {
	return addRecord(ctx, r.DB, m)
}

func (r *Repository[T]) BulkAddContext(ctx context.Context, m []T) error {
	baseModels := make([]BaseModel, len(m))
	for i, model := range m {
		baseModels[i] = model
	}
	return bulkAddRecords(ctx, r.DB, baseModels)
}

func (r *Repository[T]) BulkContext(ctx context.Context, ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	m := r.New()
	return bulkRecords(ctx, r.DB, m.Schema(), m.TableName(), m.PrimaryKey(), fields, ids, limit, pageCursor, orderBy, order)
}

func (r *Repository[T]) DeadDetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(ctx, r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, true)
}

func (r *Repository[T]) DeadListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(ctx, r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, true)
}

func (r *Repository[T]) DeleteContext(ctx context.Context, m T) error {
	return deleteRecord(ctx, r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}

func (r *Repository[T]) DetailContext(ctx context.Context, id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getRecord(ctx, r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, false)
}

func (r *Repository[T]) EditContext(ctx context.Context, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	return editRecord(ctx, r.DB, table, pk, pkVal, cols, vals)
}

func (r *Repository[T]) ListContext(ctx context.Context, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listRecords(ctx, r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, false)
}

func (r *Repository[T]) ListOneContext(ctx context.Context, orderBy, order string, fields []string, filters []helper.Filter) (map[string]any, error) {
	results, err := r.ListContext(ctx, 1, nil, orderBy, order, fields, filters)
	if len(results) == 0 {
		return make(map[string]any), err
	}
	return results[0], err
}

func (r *Repository[T]) RawContext(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	m := r.New()
	sqlText, args := helper.PrepareRawQuery(query, params)
	return rawRecords(ctx, r.DB, m.Schema(), sqlText, args...)
}

func (r *Repository[T]) UndeleteContext(ctx context.Context, m T) error {
	return undeleteRecord(ctx, r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}
