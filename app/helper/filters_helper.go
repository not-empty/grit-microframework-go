package helper

import (
	"fmt"
	"net/http"
	"strings"
)

type Filter struct {
	Field    string
	Operator string
	Value    string
}

func GetFilters(r *http.Request, allowed []string) (filters []Filter) {
	query := r.URL.Query()

	allowedMap := make(map[string]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	for _, raw := range query["filter"] {
		parts := strings.SplitN(raw, ":", 3)
		if len(parts) != 3 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		operator := strings.TrimSpace(strings.ToLower(parts[1]))
		value := strings.TrimSpace(parts[2])

		if _, ok := allowedMap[field]; ok {
			filters = append(filters, Filter{
				Field:    field,
				Operator: operator,
				Value:    value,
			})
		}
	}

	return filters
}

func BuildWhereClause(filters []Filter) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	for _, f := range filters {
		switch f.Operator {
		case "eql":
			clauses = append(clauses, fmt.Sprintf("%s = ?", f.Field))
			args = append(args, f.Value)
		case "neq":
			clauses = append(clauses, fmt.Sprintf("%s != ?", f.Field))
			args = append(args, f.Value)
		case "lik":
			clauses = append(clauses, fmt.Sprintf("%s LIKE ?", f.Field))
			args = append(args, "%"+f.Value+"%")
		case "gt":
			clauses = append(clauses, fmt.Sprintf("%s > ?", f.Field))
			args = append(args, f.Value)
		case "lt":
			clauses = append(clauses, fmt.Sprintf("%s < ?", f.Field))
			args = append(args, f.Value)
		case "gte":
			clauses = append(clauses, fmt.Sprintf("%s >= ?", f.Field))
			args = append(args, f.Value)
		case "lte":
			clauses = append(clauses, fmt.Sprintf("%s <= ?", f.Field))
			args = append(args, f.Value)
		case "btw":
			rangeParts := strings.Split(f.Value, ",")
			if len(rangeParts) == 2 {
				clauses = append(clauses, fmt.Sprintf("%s BETWEEN ? AND ?", f.Field))
				args = append(args, rangeParts[0], rangeParts[1])
			}
		case "nul":
			if f.Value == "true" {
				clauses = append(clauses, fmt.Sprintf("%s IS NULL", f.Field))
			} else {
				clauses = append(clauses, fmt.Sprintf("%s IS NOT NULL", f.Field))
			}
		case "nnu":
			clauses = append(clauses, fmt.Sprintf("%s IS NOT NULL", f.Field))
		case "in":
			inParts := strings.Split(f.Value, ",")
			if len(inParts) > 0 {
				placeholders := strings.Repeat("?,", len(inParts))
				placeholders = strings.TrimRight(placeholders, ",")
				clauses = append(clauses, fmt.Sprintf("%s IN (%s)", f.Field, placeholders))
				for _, val := range inParts {
					args = append(args, strings.TrimSpace(val))
				}
			}
		}
	}

	if len(clauses) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}
