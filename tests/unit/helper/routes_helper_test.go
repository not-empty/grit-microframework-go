package helper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

type mockModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (m *mockModel) Sanitize() {
	m.Name = "sanitized"
}

func TestJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"msg": "ok"}

	helper.JSONResponse(rec, http.StatusCreated, payload)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var data map[string]string
	err := json.NewDecoder(rec.Body).Decode(&data)
	require.NoError(t, err)
	require.Equal(t, "ok", data["msg"])
}

func TestExtractID_Valid(t *testing.T) {
	id, err := helper.ExtractID("/users/123", "/users/")
	require.NoError(t, err)
	require.Equal(t, "123", id)
}

func TestExtractID_Missing(t *testing.T) {
	_, err := helper.ExtractID("/users/", "/users/")
	require.Error(t, err)
	require.Equal(t, "Missing ID", err.Error())
}

func TestFilterList(t *testing.T) {
	data := []mockModel{
		{ID: "1", Name: "A"},
		{ID: "2", Name: "B"},
	}
	fields := []string{"id"}

	result := helper.FilterList(data, fields)
	require.Len(t, result, 2)
	require.Equal(t, "1", result[0]["id"])
	require.Nil(t, result[0]["name"])
}

func TestSanitizeModel_WithSanitize(t *testing.T) {
	model := &mockModel{ID: "1", Name: "Original"}
	helper.SanitizeModel(model)
	require.Equal(t, "sanitized", model.Name)
}

func TestSanitizeModel_NoSanitize(t *testing.T) {
	data := map[string]string{"key": "value"}
	helper.SanitizeModel(data)
}
