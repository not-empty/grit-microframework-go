package helper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestGetPaginationParams_Default(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	limit, offset, err := helper.GetPaginationParams(req)

	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 0, offset)
}

func TestGetPaginationParams_Page2(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "page=2"}}
	limit, offset, err := helper.GetPaginationParams(req)

	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 5, offset)
}

func TestGetPaginationParams_InvalidPage(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "page=abc"}}
	limit, offset, err := helper.GetPaginationParams(req)

	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 0, offset)
}

func TestGetPaginationParams_ZeroPage(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "page=0"}}
	limit, offset, err := helper.GetPaginationParams(req)

	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 0, offset)
}

func TestGetPaginationParams_NegativePage(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "page=-3"}}
	limit, offset, err := helper.GetPaginationParams(req)

	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 0, offset)
}
