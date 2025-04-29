package models

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/not-empty/grit/app/helper"
	"time"
)

type Example struct {
	ID        string           `json:"id"`
	Name      string           `json:"name" validate:"required,min=5"`
	Age       int              `json:"age" validate:"required,number,gt=0,lt=100"`
	LastLogin *helper.JSONTime `json:"last_login"`
	CreatedAt *time.Time       `json:"created_at"`
	UpdatedAt *time.Time       `json:"updated_at"`
	DeletedAt *time.Time       `json:"deleted_at"`
}

func (m *Example) Schema() map[string]string {
	return map[string]string{
		"id":         "string",
		"name":       "string",
		"age":        "int",
		"last_login": "*time.Time",
		"created_at": "*time.Time",
		"updated_at": "*time.Time",
		"deleted_at": "*time.Time",
	}
}

func (m *Example) TableName() string {
	return "example"
}

func (m *Example) Columns() []string {
	return []string{"id", "name", "age", "last_login", "created_at", "updated_at", "deleted_at"}
}

func (m *Example) Values() []interface{} {
	return []interface{}{m.ID, m.Name, m.Age, m.LastLogin, m.CreatedAt, m.UpdatedAt, m.DeletedAt}
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
	policy := bluemonday.UGCPolicy()
	m.Name = policy.Sanitize(m.Name)
}
