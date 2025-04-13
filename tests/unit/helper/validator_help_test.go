package helper_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func TestValidatePayload_Valid(t *testing.T) {
	helper.InjectValidator(validator.New())

	ts := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	w := httptest.NewRecorder()
	err := helper.ValidatePayload(w, ts)

	require.NoError(t, err)
	require.Equal(t, 200, w.Result().StatusCode)
}

func TestValidatePayload_Invalid(t *testing.T) {
	helper.InjectValidator(validator.New())

	ts := TestStruct{
		Name:  "",
		Email: "invalid-email",
	}

	w := httptest.NewRecorder()
	err := helper.ValidatePayload(w, ts)

	require.Error(t, err)
	require.Equal(t, 422, w.Result().StatusCode)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	require.Contains(t, response, "errors")
	require.Greater(t, len(response["errors"].([]interface{})), 0)
}

func TestValidatePayload_UnexpectedError(t *testing.T) {
	w := httptest.NewRecorder()
	// simulate a type that causes an unexpected error
	invalidInput := make(chan int) // validator can't validate this

	err := helper.ValidatePayload(w, invalidInput)

	require.Error(t, err)
	require.Contains(t, w.Body.String(), "errors")
}
