package models

import (
	"time"
)

type Example struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (m *Example) TableName() string {
	return "example"
}

func (m *Example) Columns() []string {
	return []string{"id", "name", "age", "created_at", "updated_at", "deleted_at"}
}

func (m *Example) Values() []interface{} {
	return []interface{}{m.ID, m.Name, m.Age, m.CreatedAt, m.UpdatedAt, m.DeletedAt}
}

func (m *Example) PrimaryKey() string {
	return "id"
}

func (m *Example) PrimaryKeyValue() interface{} {
	return m.ID
}

func (m *Example) SetCreatedAt(t time.Time) {
	m.CreatedAt = &t
}

func (m *Example) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = &t
}

func (m *Example) Sanitize() {
	// no fields to sanitize
}
