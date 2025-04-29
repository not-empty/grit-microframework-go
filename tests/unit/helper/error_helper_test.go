package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/not-empty/grit/app/config"
	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestJSONErrorSimple_WithString(t *testing.T) {
	os.Setenv("APP_ENV", "local")

	_ = config.LoadConfig()

	rec := httptest.NewRecorder()
	helper.JSONErrorSimple(rec, http.StatusNotFound, "not found")

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "not found", resp.Error)
	require.Equal(t, "", resp.Detail)
}

func TestJSONError_WithError_Prod(t *testing.T) {
	os.Setenv("APP_ENV", "prod")

	_ = config.LoadConfig()

	rec := httptest.NewRecorder()
	mes := "error"
	err := errors.New("something went wrong")

	helper.JSONError(rec, http.StatusBadRequest, mes, err)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err2 := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err2)
	require.Equal(t, "error", resp.Error)
	require.Equal(t, "", resp.Detail)
}

func TestJSONError_WithError_Local(t *testing.T) {
	os.Setenv("APP_ENV", "local")

	_ = config.LoadConfig()

	rec := httptest.NewRecorder()
	mes := "error"
	err := errors.New("something went wrong")

	helper.JSONError(rec, http.StatusBadRequest, mes, err)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp helper.ErrorResponse
	err2 := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err2)
	require.Equal(t, "error", resp.Error)
	require.Equal(t, "something went wrong", resp.Detail)
}
