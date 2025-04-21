package repository

import (
	"database/sql"
	"time"

	"github.com/not-empty/grit/app/helper"
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
	PrimaryKey() string
	PrimaryKeyValue() interface{}
	Schema() map[string]string
}

type RepositoryInterface[T BaseModel] interface {
	New() T
	Insert(m T) error
	UpdateFields(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error
	Delete(m T) error
	Get(id interface{}, fields []string) (map[string]any, error)
	GetDeleted(id interface{}, fields []string) (map[string]any, error)
	ListActive(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	ListDeleted(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error)
	BulkGet(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error)
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

func (r *Repository[T]) Insert(m T) error {
	return insertModel(r.DB, m)
}

func (r *Repository[T]) UpdateFields(table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	return updateModelFields(r.DB, table, pk, pkVal, cols, vals)
}

func (r *Repository[T]) Delete(m T) error {
	return softDeleteModel(r.DB, m.TableName(), m.PrimaryKey(), m.PrimaryKeyValue())
}

func (r *Repository[T]) Get(id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getModel(r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, false)
}

func (r *Repository[T]) GetDeleted(id interface{}, fields []string) (map[string]any, error) {
	m := r.New()
	return getModel(r.DB, id, m.Schema(), m.TableName(), m.PrimaryKey(), fields, true)
}

func (r *Repository[T]) ListActive(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listModels(r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, false)
}

func (r *Repository[T]) ListDeleted(limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string, filters []helper.Filter) ([]map[string]any, error) {
	m := r.New()
	return listModels(r.DB, m.Schema(), m.TableName(), fields, limit, pageCursor, orderBy, order, filters, true)
}

func (r *Repository[T]) BulkGet(ids []string, limit int, pageCursor *helper.PageCursor, orderBy, order string, fields []string) ([]map[string]any, error) {
	m := r.New()
	return bulkGetModels(r.DB, m.Schema(), m.TableName(), m.PrimaryKey(), fields, ids, limit, pageCursor, orderBy, order)
}
