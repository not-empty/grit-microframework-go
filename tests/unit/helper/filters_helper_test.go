package helper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/not-empty/grit-microframework-go/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetFilters_Valid(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "filter=name:eql:John&filter=age:gt:30",
		},
	}
	allowed := []string{"name", "age"}

	result := helper.GetFilters(req, allowed)
	require.Len(t, result, 2)
	require.Equal(t, "name", result[0].Field)
	require.Equal(t, "eql", result[0].Operator)
	require.Equal(t, "John", result[0].Value)
}

func TestGetFilters_InvalidSyntax(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "filter=invalid",
		},
	}
	allowed := []string{"name"}

	result := helper.GetFilters(req, allowed)
	require.Empty(t, result)
}

func TestGetFilters_NotAllowed(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "filter=secret:eql:val",
		},
	}
	allowed := []string{"name"}

	result := helper.GetFilters(req, allowed)
	require.Empty(t, result)
}

func TestBuildWhereClause_AllOperators(t *testing.T) {
	filters := []helper.Filter{
		{Field: "name", Operator: "eql", Value: "John"},
		{Field: "age", Operator: "gt", Value: "30"},
		{Field: "status", Operator: "neq", Value: "inactive"},
		{Field: "title", Operator: "lik", Value: "engineer"},
		{Field: "score", Operator: "lt", Value: "90"},
		{Field: "score", Operator: "gte", Value: "60"},
		{Field: "score", Operator: "lte", Value: "100"},
		{Field: "created", Operator: "btw", Value: "2020-01-01,2020-12-31"},
		{Field: "deleted_at", Operator: "nul", Value: "true"},
		{Field: "updated_at", Operator: "nul", Value: "false"},
		{Field: "email", Operator: "nnu", Value: ""},
		{Field: "role", Operator: "in", Value: "admin,user"},
	}

	where, args := helper.BuildWhereClause(filters)
	require.Contains(t, where, "`name` = ?")
	require.Contains(t, where, "`age` > ?")
	require.Contains(t, where, "`status` != ?")
	require.Contains(t, where, "`title` LIKE ?")
	require.Contains(t, where, "`score` < ?")
	require.Contains(t, where, "`score` >= ?")
	require.Contains(t, where, "`score` <= ?")
	require.Contains(t, where, "`created` BETWEEN ? AND ?")
	require.Contains(t, where, "`deleted_at` IS NULL")
	require.Contains(t, where, "`updated_at` IS NOT NULL")
	require.Contains(t, where, "`email` IS NOT NULL")
	require.Contains(t, where, "`role` IN (?,?)")
	require.Len(t, args, 11)
}

func TestBuildWhereClause_Empty(t *testing.T) {
	where, args := helper.BuildWhereClause(nil)
	require.Equal(t, "", where)
	require.Empty(t, args)
}
