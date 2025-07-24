package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const DefaultPageLimit = 25

type PageCursor struct {
	LastID    string `json:"last_id"`
	LastValue string `json:"last_value"`
}

func EncodeCursor(c PageCursor) string {
	data, _ := json.Marshal(c)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}

func DecodeCursor(token string) (PageCursor, error) {
	var c PageCursor
	raw, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(token)
	if err != nil {
		return c, errors.New("invalid cursor encoding")
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, errors.New("invalid cursor payload")
	}
	return c, nil
}

func GetPaginationParams(r *http.Request) (limit int, cursor *PageCursor, err error) {
	limit = DefaultPageLimit

	raw := r.URL.Query().Get("page_cursor")
	if raw == "" {
		return
	}

	var c PageCursor
	rawBytes, decErr := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(raw)
	if decErr != nil {
		err = errors.New("invalid cursor encoding")
		return
	}
	if jsonErr := json.Unmarshal(rawBytes, &c); jsonErr != nil {
		err = errors.New("invalid cursor payload")
		return
	}
	cursor = &c
	return
}

func BuildPageCursor(body []byte, query url.Values) (string, error) {
	if len(body) == 0 || body[0] != '[' {
		return "", nil
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil || len(arr) == 0 {
		return "", nil
	}

	if len(arr) < DefaultPageLimit {
		return "", nil
	}

	last := arr[len(arr)-1]
	orderBy := query.Get("order_by")
	if orderBy == "" {
		orderBy = "id"
	}

	idVal, _ := last["id"].(string)
	var lastVal string
	if v, ok := last[orderBy]; ok {
		lastVal = fmt.Sprintf("%v", v)
	}

	return EncodeCursor(PageCursor{
		LastID:    idVal,
		LastValue: lastVal,
	}), nil
}
