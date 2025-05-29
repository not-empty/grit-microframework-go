package helper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeCursor_Roundtrip(t *testing.T) {
	orig := helper.PageCursor{LastID: "abc123", LastValue: "xyz789"}
	token := helper.EncodeCursor(orig)

	decoded, err := helper.DecodeCursor(token)
	require.NoError(t, err)
	require.Equal(t, orig.LastID, decoded.LastID)
	require.Equal(t, orig.LastValue, decoded.LastValue)
}

func TestDecodeCursor_InvalidEncoding(t *testing.T) {
	_, err := helper.DecodeCursor("!!!not-base64$$$")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid cursor encoding")
}

func TestDecodeCursor_InvalidPayload(t *testing.T) {
	bad := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte("not-json"))
	_, err := helper.DecodeCursor(bad)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid cursor payload")
}

func TestGetPaginationParams_NoCursor(t *testing.T) {
	req := httptest.NewRequest("GET", "/?foo=bar", nil)
	limit, cursor, err := helper.GetPaginationParams(req)
	require.NoError(t, err)
	require.Equal(t, helper.DefaultPageLimit, limit)
	require.Nil(t, cursor)
}

func TestGetPaginationParams_ValidCursor(t *testing.T) {
	orig := helper.PageCursor{LastID: "id42", LastValue: "val42"}
	tok := helper.EncodeCursor(orig)

	req := httptest.NewRequest("GET", "/?page_cursor="+tok, nil)
	limit, cursor, err := helper.GetPaginationParams(req)
	require.NoError(t, err)
	require.Equal(t, helper.DefaultPageLimit, limit)
	require.NotNil(t, cursor)
	require.Equal(t, orig.LastID, cursor.LastID)
	require.Equal(t, orig.LastValue, cursor.LastValue)
}

func TestGetPaginationParams_InvalidCursorEncoding(t *testing.T) {
	req := httptest.NewRequest("GET", "/?page_cursor=not-base64!", nil)

	limit, cursor, err := helper.GetPaginationParams(req)

	require.Error(t, err)
	require.EqualError(t, err, "Invalid cursor encoding")

	require.Nil(t, cursor)
	require.Equal(t, helper.DefaultPageLimit, limit)
}

func TestGetPaginationParams_InvalidCursorPayload(t *testing.T) {
	badJSON := base64.URLEncoding.WithPadding(base64.NoPadding).
		EncodeToString([]byte("not-json"))

	req := httptest.NewRequest("GET", "/?page_cursor="+badJSON, nil)

	limit, cursor, err := helper.GetPaginationParams(req)

	require.Error(t, err)
	require.EqualError(t, err, "Invalid cursor payload")
	require.Nil(t, cursor)
	require.Equal(t, helper.DefaultPageLimit, limit)
}

func TestBuildPageCursor_NonArrayAndInvalidJSON(t *testing.T) {
	c, err := helper.BuildPageCursor([]byte(`{"foo":1}`), url.Values{})
	require.NoError(t, err)
	require.Empty(t, c)

	c, err = helper.BuildPageCursor([]byte(`[1,2`), url.Values{})
	require.NoError(t, err)
	require.Empty(t, c)
}

func TestBuildPageCursor_ShorterThanLimit(t *testing.T) {
	arr := make([]map[string]interface{}, helper.DefaultPageLimit-1)
	for i := range arr {
		arr[i] = map[string]interface{}{"id": fmt.Sprintf("id%d", i+1), "score": float64(i * 10)}
	}
	body, err := json.Marshal(arr)
	require.NoError(t, err)

	cursor, err := helper.BuildPageCursor(body, url.Values{"order_by": []string{"score"}})
	require.NoError(t, err)
	require.Empty(t, cursor)
}

func TestBuildPageCursor_FullPage_DefaultOrderBy(t *testing.T) {
	arr := make([]map[string]interface{}, helper.DefaultPageLimit)
	for i := range arr {
		arr[i] = map[string]interface{}{
			"id":    fmt.Sprintf("id%d", i+1),
			"score": float64((i + 1) * 100),
		}
	}
	body, err := json.Marshal(arr)
	require.NoError(t, err)

	cursor, err := helper.BuildPageCursor(body, url.Values{})
	require.NoError(t, err)
	require.NotEmpty(t, cursor)

	pc, err := helper.DecodeCursor(cursor)
	require.NoError(t, err)
	require.Equal(t, "id25", pc.LastID)
	require.Equal(t, "id25", pc.LastValue)
}

func TestBuildPageCursor_FullPage_CustomOrderBy(t *testing.T) {
	arr := make([]map[string]interface{}, helper.DefaultPageLimit)
	for i := range arr {
		arr[i] = map[string]interface{}{
			"id":    fmt.Sprintf("id%d", i+1),
			"score": float64((i + 1) * 7),
		}
	}
	body, err := json.Marshal(arr)
	require.NoError(t, err)

	query := url.Values{"order_by": []string{"score"}}
	cursor, err := helper.BuildPageCursor(body, query)
	require.NoError(t, err)

	pc, err := helper.DecodeCursor(cursor)
	require.NoError(t, err)
	require.Equal(t, "id25", pc.LastID)
	require.Equal(t, "175", pc.LastValue)
}
