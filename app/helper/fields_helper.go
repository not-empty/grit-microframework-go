package helper

import (
	"net/http"
	"strings"
)

func GetFieldsParam(r *http.Request, allowedFields []string) []string {
	query := r.URL.Query().Get("fields")
	if query == "" {
		return nil
	}

	requested := strings.Split(query, ",")
	allowedMap := make(map[string]struct{}, len(allowedFields))
	for _, f := range allowedFields {
		allowedMap[f] = struct{}{}
	}

	var valid []string
	for _, field := range requested {
		field = strings.TrimSpace(field)
		if _, ok := allowedMap[field]; ok {
			valid = append(valid, field)
		}
	}

	if len(valid) == 0 {
		return nil
	}
	return valid
}

func FilterFields(requested, allowed []string) []string {
	if len(requested) == 0 {
		return allowed
	}

	allowedMap := make(map[string]bool, len(allowed))
	for _, col := range allowed {
		allowedMap[col] = true
	}

	var filtered []string
	for _, col := range requested {
		if allowedMap[col] {
			filtered = append(filtered, col)
		}
	}

	if len(filtered) == 0 {
		return allowed
	}
	return filtered
}

func ValidateOrder(order string) string {
	if order != "ASC" && order != "DESC" {
		return "DESC"
	}
	return order
}

func ValidateOrderBy(orderBy string, allowed []string) string {
	allowedMap := make(map[string]bool, len(allowed))
	for _, col := range allowed {
		allowedMap[col] = true
	}
	if allowedMap[orderBy] {
		return orderBy
	}
	return "id"
}
