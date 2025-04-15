package helper

import (
	"net/http"
	"strconv"
)

func GetPaginationParams(r *http.Request) (limit int, offset int, err error) {
	const fixedLimit = 5
	page := 1

	query := r.URL.Query()
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit = fixedLimit
	offset = (page - 1) * fixedLimit
	return limit, offset, nil
}
