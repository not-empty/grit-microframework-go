package helper

import (
	"fmt"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestRegisterGetRawQuery(t *testing.T) {
	table, name, sql := "tbl", "q1", "SELECT * FROM tbl"
	got, ok := helper.GetRawQuery(table, name)
	require.False(t, ok)
	require.Empty(t, got)

	helper.RegisterRawQueries(table, map[string]string{name: sql})
	got2, ok2 := helper.GetRawQuery(table, name)
	require.True(t, ok2)
	require.Equal(t, sql, got2)
}

func TestCheckRawQueryAllowed_DenySubstring(t *testing.T) {
	deny := []string{";", "--", "/*", "*/"}
	for _, bad := range deny {
		q := fmt.Sprintf("SELECT 1 %s", bad)
		ok, err := helper.CheckRawQueryAllowed(q)
		require.False(t, ok)
		require.EqualError(t, err, fmt.Sprintf("forbidden substring in query: %s", bad))
	}
}

func TestCheckRawQueryAllowed_DenyWordAndAllowEmbedded(t *testing.T) {
	ok, err := helper.CheckRawQueryAllowed("SELECT * FROM tbl WHERE delete = 1")
	require.False(t, ok)
	require.EqualError(t, err, "forbidden keyword in query: delete")

	ok2, err2 := helper.CheckRawQueryAllowed("SELECT created_at FROM tbl")
	require.True(t, ok2)
	require.NoError(t, err2)
}

func TestCheckRawQueryAllowed_AllowPrefixes(t *testing.T) {
	cases := []string{
		"SELECT id FROM tbl",
		"  select 1",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
	}
	for _, q := range cases {
		ok, err := helper.CheckRawQueryAllowed(q)
		require.True(t, ok)
		require.NoError(t, err)
	}
}

func TestCheckRawQueryAllowed_RejectOther(t *testing.T) {
	ok, err := helper.CheckRawQueryAllowed("SHOW TABLES")
	require.False(t, ok)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only [select with] queries are allowed")
}

func TestExtractRawParams(t *testing.T) {
	q := "SELECT * FROM tbl WHERE a=:a AND b= :b OR c=:a"
	exp := []string{"a", "b"}
	got := helper.ExtractRawParams(q)
	require.Equal(t, exp, got)

	require.Empty(t, helper.ExtractRawParams("SELECT 1"))
}

func TestValidateRawParams(t *testing.T) {
	q := "SELECT * FROM tbl WHERE x=:x AND y=:y"
	require.NoError(t, helper.ValidateRawParams(q, map[string]any{"x": 1, "y": 2}))
	err := helper.ValidateRawParams(q, map[string]any{"x": 1})
	require.EqualError(t, err, "missing parameter: y")
	err2 := helper.ValidateRawParams(q, map[string]any{"x": 1, "y": 2, "z": 3})
	require.EqualError(t, err2, "unexpected parameter: z")
}

func TestPrepareRawQuery(t *testing.T) {
	q := "SELECT * FROM tbl WHERE a=:a AND b=:b"
	params := map[string]any{"a": 1, "b": 2}
	expSQL := "SELECT * FROM tbl WHERE a=? AND b=? LIMIT 25"
	sql, args := helper.PrepareRawQuery(q, params)
	require.Equal(t, expSQL, sql)
	require.Equal(t, []interface{}{1, 2}, args)

	q2 := "SELECT * FROM tbl"
	exp2 := "SELECT * FROM tbl LIMIT 25"
	sql2, args2 := helper.PrepareRawQuery(q2, nil)
	require.Equal(t, exp2, sql2)
	require.Empty(t, args2)

	q3 := "SELECT * FROM tbl WHERE c=:c OR d=:c"
	params3 := map[string]any{"c": 3}
	expSQL3 := "SELECT * FROM tbl WHERE c=? OR d=? LIMIT 25"
	sql3, args3 := helper.PrepareRawQuery(q3, params3)
	require.Equal(t, expSQL3, sql3)
	require.Equal(t, []interface{}{3, 3}, args3)
}
