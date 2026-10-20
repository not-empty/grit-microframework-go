package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const DefaultPageLimit = 25

type PageCursor struct {
	LastID        string `json:"last_id"`
	LastValue     string `json:"last_value"`
	LastValueNull bool   `json:"last_value_null,omitempty"`
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
	limit = ParseLimit(r.URL.Query().Get("limit"))

	raw := r.URL.Query().Get("page_cursor")
	if raw == "" {
		return
	}

	var c PageCursor
	rawBytes, decErr := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(raw)
	if decErr != nil {
		err = errors.New("Invalid cursor encoding")
		return
	}
	if jsonErr := json.Unmarshal(rawBytes, &c); jsonErr != nil {
		err = errors.New("Invalid cursor payload")
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

	lim := ParseLimit(query.Get("limit"))
	if len(arr) < lim {
		return "", nil
	}

	last := arr[len(arr)-1]
	orderBy := query.Get("order_by")
	if orderBy == "" {
		orderBy = "id"
	}

	idVal, _ := last["id"].(string)
	var lastVal string
	var lastValNull bool
	if v, ok := last[orderBy]; ok {
		if v == nil {
			lastValNull = true
		} else {
			lastVal = fmt.Sprintf("%v", v)
		}
	}

	return EncodeCursor(PageCursor{
		LastID:        idVal,
		LastValue:     lastVal,
		LastValueNull: lastValNull,
	}), nil
}

func ParseLimit(raw string) int {
	def := DefaultPageLimit
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 || v >= def {
		return def
	}
	return v
}
