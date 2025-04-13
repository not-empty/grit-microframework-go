package helper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetOrderParams_AllValid(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "order_by=name&order=asc"}}
	orderBy, orderDir := helper.GetOrderParams(req, "id")

	require.Equal(t, "name", orderBy)
	require.Equal(t, "asc", orderDir)
}

func TestGetOrderParams_MissingOrderBy(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "order=asc"}}
	orderBy, orderDir := helper.GetOrderParams(req, "id")

	require.Equal(t, "id", orderBy)
	require.Equal(t, "asc", orderDir)
}

func TestGetOrderParams_InvalidOrderDirection(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "order_by=created_at&order=random"}}
	orderBy, orderDir := helper.GetOrderParams(req, "id")

	require.Equal(t, "created_at", orderBy)
	require.Equal(t, "desc", orderDir) // fallback
}

func TestGetOrderParams_EmptyQuery(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	orderBy, orderDir := helper.GetOrderParams(req, "id")

	require.Equal(t, "id", orderBy)
	require.Equal(t, "desc", orderDir)
}
