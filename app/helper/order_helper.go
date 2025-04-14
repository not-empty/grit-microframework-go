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

	orderDir = strings.ToUpper(query.Get("order"))
	if orderDir != "ASC" && orderDir != "DESC" {
		orderDir = "DESC"
	}

	return
}
