package helper

import (
	"net/http"
	"strings"
)

func GetFieldsParam(r *http.Request, allowedFields []string) (fields []string) {
	query := r.URL.Query().Get("fields")
	if query == "" {
		return nil
	}

	requested := strings.Split(query, ",")
	allowedMap := make(map[string]struct{}, len(allowedFields))
	for _, f := range allowedFields {
		allowedMap[f] = struct{}{}
	}

	for _, field := range requested {
		field = strings.TrimSpace(field)
		if _, ok := allowedMap[field]; ok {
			fields = append(fields, field)
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
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

func EscapeMysqlField(field string) string {
	return "`" + field + "`"
}

func EscapeMysqlFields(fields []string) []string {
	escapedFields := make([]string, len(fields))
	for i, field := range fields {
		escapedFields[i] = EscapeMysqlField(field)
	}
	return escapedFields
}

func EnsurePaginationFields(fields []string, orderBy string) []string {
	if len(fields) == 0 {
		return fields
	}

	seen := make(map[string]struct{}, len(fields))
	result := make([]string, 0, len(fields)+2)

	addUnique := func(field string) {
		if _, exists := seen[field]; !exists {
			seen[field] = struct{}{}
			result = append(result, field)
		}
	}

	for _, f := range fields {
		addUnique(f)
	}

	addUnique("id")
	if orderBy != "" && orderBy != "id" {
		addUnique(orderBy)
	}

	return result
}
