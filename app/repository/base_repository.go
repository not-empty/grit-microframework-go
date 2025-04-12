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
}

type Repository[T BaseModel] struct {
	DB  *sql.DB
	New func() T
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

func (r *Repository[T]) Get(id interface{}, m T, fields []string) error {
	return getModel(r.DB, id, m, fields, false)
}

func (r *Repository[T]) GetDeleted(id interface{}, fields []string) (T, error) {
	m := r.New()
	err := getModel(r.DB, id, m, fields, true)
	return m, err
}

func (r *Repository[T]) ListActive(limit, offset int, orderBy, order string, fields []string, filters []helper.Filter) ([]T, error) {
	return listModels(r.DB, r.New, r.New().TableName(), r.New().Columns(), fields, limit, offset, orderBy, order, filters, false)
}

func (r *Repository[T]) ListDeleted(limit, offset int, orderBy, order string, fields []string) ([]T, error) {
	return listModels(r.DB, r.New, r.New().TableName(), r.New().Columns(), fields, limit, offset, orderBy, order, nil, true)
}

func (r *Repository[T]) BulkGet(ids []string, limit, offset int, orderBy, order string, fields []string) ([]T, error) {
	return bulkGetModels(r.DB, r.New, r.New().TableName(), r.New().Columns(), fields, ids, limit, offset, orderBy, order)
}
