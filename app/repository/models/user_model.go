package models

import (
	"github.com/microcosm-cc/bluemonday"
	"time"
)

type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (m *User) Schema() map[string]string {
	return map[string]string{
		"id":         "string",
		"name":       "string",
		"email":      "string",
		"created_at": "*time.Time",
		"updated_at": "*time.Time",
		"deleted_at": "*time.Time",
	}
}

func (m *User) TableName() string {
	return "user"
}

func (m *User) Columns() []string {
	return []string{"id", "name", "email", "created_at", "updated_at", "deleted_at"}
}

func (m *User) Values() []interface{} {
	return []interface{}{m.ID, m.Name, m.Email, m.CreatedAt, m.UpdatedAt, m.DeletedAt}
}

func (m *User) PrimaryKey() string {
	return "id"
}

func (m *User) PrimaryKeyValue() interface{} {
	return m.ID
}

func (m *User) SetCreatedAt(t time.Time) {
	m.CreatedAt = &t
}

func (m *User) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = &t
}

func (m *User) Sanitize() {
	policy := bluemonday.UGCPolicy()
	m.Name = policy.Sanitize(m.Name)
}
