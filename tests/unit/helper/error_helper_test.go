package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestJSONError_WithError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := errors.New("something went wrong")

	helper.JSONError(rec, http.StatusBadRequest, err)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err2 := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err2)
	require.Equal(t, "something went wrong", resp.Error)
}

func TestJSONError_WithString(t *testing.T) {
	rec := httptest.NewRecorder()
	helper.JSONError(rec, http.StatusNotFound, "not found")

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "not found", resp.Error)
}

func TestJSONError_WithUnknownType(t *testing.T) {
	rec := httptest.NewRecorder()
	helper.JSONError(rec, http.StatusInternalServerError, 12345) // unsupported type

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "Unknown error", resp.Error)
}
