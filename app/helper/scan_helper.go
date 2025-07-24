package helper

import (
	"database/sql"
	"fmt"
	"strings"
)

type RowScanner interface {
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type rowsAdapter struct {
	*sql.Rows
}

func NewRowsAdapter(r *sql.Rows) RowScanner {
	return &rowsAdapter{r}
}

func (r *rowsAdapter) Columns() ([]string, error) {
	return r.Rows.Columns()
}

func (r *rowsAdapter) ColumnTypes() ([]*sql.ColumnType, error) {
	return r.Rows.ColumnTypes()
}

func (r *rowsAdapter) Next() bool {
	return r.Rows.Next()
}

func (r *rowsAdapter) Scan(dest ...any) error {
	return r.Rows.Scan(dest...)
}

func (r *rowsAdapter) Err() error {
	return r.Rows.Err()
}

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
	isDate := make(map[string]struct{})

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
		case "*time.time", "time.time":
			ptr := new(sql.NullTime)
			scanMap[col] = ptr
			scanArgs[i] = ptr
		case "date":
			ptr := new(sql.NullTime)
			scanMap[col] = ptr
			scanArgs[i] = ptr
			isDate[col] = struct{}{}
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
				_, ok := isDate[key]
				if ok {
					result[key] = v.Time.Format("2006-01-02")
					continue
				}

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

func SimpleScanRows(rs RowScanner) ([]map[string]any, error) {
	cols, _ := rs.Columns()

	var results []map[string]any
	for rs.Next() {
		values := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rs.Scan(scanArgs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any, len(cols))
		for i, col := range cols {
			v := values[i]
			switch x := v.(type) {
			case nil:
				rowMap[col] = nil
			case []byte:
				rowMap[col] = string(x)
			default:
				rowMap[col] = x
			}
		}
		results = append(results, rowMap)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
