package helper

import (
	"database/sql"
	"fmt"
	"strings"
)

func GenericScanToMap(scanner interface {
	Columns() ([]string, error)
	Scan(...any) error
}, schema map[string]string) (map[string]any, error) {
	cols, err := scanner.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	scanMap := make(map[string]any)
	scanArgs := make([]any, len(cols))

	for i, col := range cols {
		typ, ok := schema[col]
		if !ok {
			var discard any
			scanArgs[i] = &discard
			continue
		}

		switch strings.ToLower(typ) {
		case "string":
			ptr := new(sql.NullString)
			scanMap[col] = ptr
			scanArgs[i] = ptr
		case "int":
			ptr := new(sql.NullInt64)
			scanMap[col] = ptr
			scanArgs[i] = ptr
		case "*time.time":
			ptr := new(sql.NullTime)
			scanMap[col] = ptr
			scanArgs[i] = ptr
		default:
			var discard any
			scanArgs[i] = &discard
		}
	}

	if err := scanner.Scan(scanArgs...); err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	result := make(map[string]any)
	for key, ptr := range scanMap {
		switch v := ptr.(type) {
		case *sql.NullString:
			if v.Valid {
				result[key] = v.String
			} else {
				result[key] = ""
			}
		case *sql.NullInt64:
			if v.Valid {
				result[key] = int(v.Int64)
			} else {
				result[key] = 0
			}
		case *sql.NullTime:
			if v.Valid {
				result[key] = v.Time.Format("2006-01-02 15:04:05")
			} else {
				result[key] = nil
			}
		}
	}

	return result, nil
}

func MapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
