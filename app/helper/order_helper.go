package helper

import (
	"net/http"
	"strings"
)

func GetOrderParams(r *http.Request, defaultColumn string) (orderBy string, orderDir string) {
	query := r.URL.Query()

	orderBy = query.Get("order_by")
	if orderBy == "" {
		orderBy = defaultColumn
	}

	orderDir = strings.ToLower(query.Get("order"))
	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "desc"
	}

	return
}
