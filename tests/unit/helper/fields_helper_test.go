package helper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetFieldsParam_ValidFields(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=name,email,created_at"}}
	allowed := []string{"name", "email", "created_at", "updated_at"}

	fields, err := helper.GetFieldsParam(req, allowed)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"name", "email", "created_at"}, fields)
}

func TestGetFieldsParam_SomeInvalidFields(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=name,invalid,email"}}
	allowed := []string{"name", "email"}

	fields, err := helper.GetFieldsParam(req, allowed)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"name", "email"}, fields)
}

func TestGetFieldsParam_AllInvalid(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=bad1,bad2"}}
	allowed := []string{"name", "email"}

	fields, err := helper.GetFieldsParam(req, allowed)
	require.NoError(t, err)
	require.Nil(t, fields)
}

func TestGetFieldsParam_EmptyQuery(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	allowed := []string{"name", "email"}

	fields, err := helper.GetFieldsParam(req, allowed)
	require.NoError(t, err)
	require.Nil(t, fields)
}

func TestFilterFields_AllValid(t *testing.T) {
	requested := []string{"name", "email"}
	allowed := []string{"name", "email", "created_at"}

	result := helper.FilterFields(requested, allowed)
	require.Equal(t, requested, result)
}

func TestFilterFields_SomeInvalid(t *testing.T) {
	requested := []string{"name", "invalid"}
	allowed := []string{"name", "email"}

	result := helper.FilterFields(requested, allowed)
	require.Equal(t, []string{"name"}, result)
}

func TestFilterFields_EmptyRequested(t *testing.T) {
	requested := []string{}
	allowed := []string{"name", "email"}

	result := helper.FilterFields(requested, allowed)
	require.Equal(t, allowed, result)
}

func TestFilterFields_AllInvalid(t *testing.T) {
	requested := []string{"bad"}
	allowed := []string{"name", "email"}

	result := helper.FilterFields(requested, allowed)
	require.Equal(t, allowed, result)
}

func TestValidateOrder_ValidAsc(t *testing.T) {
	require.Equal(t, "ASC", helper.ValidateOrder("ASC"))
}

func TestValidateOrder_ValidDesc(t *testing.T) {
	require.Equal(t, "DESC", helper.ValidateOrder("DESC"))
}

func TestValidateOrder_Invalid(t *testing.T) {
	require.Equal(t, "DESC", helper.ValidateOrder("RANDOM"))
}

func TestValidateOrderBy_Allowed(t *testing.T) {
	allowed := []string{"name", "email"}
	require.Equal(t, "email", helper.ValidateOrderBy("email", allowed))
}

func TestValidateOrderBy_NotAllowed(t *testing.T) {
	allowed := []string{"name", "email"}
	require.Equal(t, "id", helper.ValidateOrderBy("created_at", allowed))
}
