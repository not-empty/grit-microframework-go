package helper

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"time"
)

var (
	typeString    = reflect.TypeOf("")
	typeInt       = reflect.TypeOf(int(0))
	typeTimePtr   = reflect.TypeOf((*time.Time)(nil))
	typeNullableT = reflect.TypeOf(sql.NullTime{})
)

func PrepareScanMap(columns []string, dest any) ([]interface{}, error) {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, errors.New("destination must be a non-nil pointer")
	}

	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, errors.New("destination must be a pointer to struct")
	}

	typ := elem.Type()
	valuePtrs := make([]interface{}, len(columns))
	columnIndex := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.ToLower(field.Tag.Get("json"))
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		columnIndex[name] = i
	}

	for i, col := range columns {
		idx, ok := columnIndex[col]
		if !ok {
			var discard any
			valuePtrs[i] = &discard
			continue
		}
		field := elem.Field(idx)
		switch field.Type() {
		case typeString:
			valuePtrs[i] = new(sql.NullString)
		case typeInt:
			valuePtrs[i] = new(sql.NullInt64)
		case typeTimePtr:
			valuePtrs[i] = new(sql.NullTime)
		default:
			var discard any
			valuePtrs[i] = &discard
		}
	}

	return valuePtrs, nil
}

func AssignScanValues(columns []string, dest any, values []interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("destination must be a non-nil pointer")
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("destination must be a pointer to struct")
	}
	typ := elem.Type()
	columnIndex := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.ToLower(field.Tag.Get("json"))
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		columnIndex[name] = i
	}
	for i, col := range columns {
		idx, ok := columnIndex[col]
		if !ok {
			continue
		}
		field := elem.Field(idx)
		src := reflect.ValueOf(values[i]).Elem().Interface()
		switch val := src.(type) {
		case sql.NullString:
			if val.Valid {
				field.SetString(val.String)
			}
		case sql.NullInt64:
			if val.Valid {
				field.SetInt(val.Int64)
			}
		case sql.NullTime:
			if val.Valid {
				t := val.Time
				field.Set(reflect.ValueOf(&t))
			}
		}
	}
	return nil
}

func GenericScanFrom[T any](scanner interface{ Columns() ([]string, error); Scan(...any) error }, target T) error {
	cols, err := scanner.Columns()
	if err != nil {
		return err
	}
	values, err := PrepareScanMap(cols, target)
	if err != nil {
		return err
	}
	if err := scanner.Scan(values...); err != nil {
		return err
	}
	return AssignScanValues(cols, target, values)
}