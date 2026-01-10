package helper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetFieldsParam_ValidFields(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=name,email,created_at"}}
	allowed := []string{"name", "email", "created_at", "updated_at"}

	fields := helper.GetFieldsParam(req, allowed)
	require.ElementsMatch(t, []string{"name", "email", "created_at"}, fields)
}

func TestGetFieldsParam_SomeInvalidFields(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=name,invalid,email"}}
	allowed := []string{"name", "email"}

	fields := helper.GetFieldsParam(req, allowed)
	require.ElementsMatch(t, []string{"name", "email"}, fields)
}

func TestGetFieldsParam_AllInvalid(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "fields=bad1,bad2"}}
	allowed := []string{"name", "email"}

	fields := helper.GetFieldsParam(req, allowed)
	require.Nil(t, fields)
}

func TestGetFieldsParam_EmptyQuery(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	allowed := []string{"name", "email"}

	fields := helper.GetFieldsParam(req, allowed)
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

func TestEscapeMysqlField(t *testing.T) {
	require.Equal(t, "`name`", helper.EscapeMysqlField("name"))
}

func TestEscapeMysqlFields(t *testing.T) {
	fields := []string{"name", "email"}

	require.Equal(t, []string{"`name`", "`email`"}, helper.EscapeMysqlFields(fields))
}

func TestEnsurePaginationFields_EmptyFields(t *testing.T) {
	fields := []string{}
	orderBy := "name"

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.Empty(t, result)
}

func TestEnsurePaginationFields_NoIdNoOrderBy(t *testing.T) {
	fields := []string{"name", "email"}
	orderBy := ""

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.ElementsMatch(t, []string{"name", "email", "id"}, result)
}

func TestEnsurePaginationFields_WithIdAlready(t *testing.T) {
	fields := []string{"id", "name", "email"}
	orderBy := ""

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.ElementsMatch(t, []string{"id", "name", "email"}, result)
}

func TestEnsurePaginationFields_WithOrderByDifferentFromId(t *testing.T) {
	fields := []string{"name"}
	orderBy := "created_at"

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.ElementsMatch(t, []string{"name", "id", "created_at"}, result)
}

func TestEnsurePaginationFields_OrderByIsId(t *testing.T) {
	fields := []string{"name", "email"}
	orderBy := "id"

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.ElementsMatch(t, []string{"name", "email", "id"}, result)
}

func TestEnsurePaginationFields_AllFieldsAlreadyPresent(t *testing.T) {
	fields := []string{"name", "id", "created_at", "email"}
	orderBy := "created_at"

	result := helper.EnsurePaginationFields(fields, orderBy)
	require.ElementsMatch(t, []string{"name", "id", "created_at", "email"}, result)
}

func TestEnsurePaginationFields_NoDuplicates(t *testing.T) {
	fields := []string{"name", "id", "email", "id"}
	orderBy := "name"

	result := helper.EnsurePaginationFields(fields, orderBy)

	idCount := 0
	nameCount := 0
	for _, f := range result {
		if f == "id" {
			idCount++
		}

		if f == "name" {
			nameCount++
		}
	}

	require.Equal(t, 1, idCount, "id should appear only once")
	require.Equal(t, 1, nameCount, "name should appear only once")
}
