package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/not-empty/grit/app/helper"
)

var ScanFunc = helper.GenericScanToMap

func addRecord(db *sql.DB, m BaseModel) error {
	cols := m.Columns()
	values := m.Values()
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.TableName(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.Exec(query, values...)
	return err
}

func bulkRecords(
	db *sql.DB,
	schema map[string]string,
	table string,
	pk string,
	fields []string,
	ids []string,
	limit int,
	pageCursor *helper.PageCursor,
	orderBy, order string,
) ([]map[string]any, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	selected := helper.FilterFields(fields, helper.MapKeys(schema))
	orderBy = helper.ValidateOrderBy(orderBy, helper.MapKeys(schema))
	order = helper.ValidateOrder(order)

	where := []string{"deleted_at IS NULL"}

	if pageCursor != nil {
		op := ">"
		if order == "DESC" {
			op = "<"
		}
		where = append(where, fmt.Sprintf(
			"(%s %s ? OR (%s = ? AND %s %s ?))",
			orderBy, op,
			orderBy, pk, op,
		))
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}
	where = append(where,
		fmt.Sprintf("%s IN (%s)", pk, strings.Join(placeholders, ", ")),
	)

	args := []interface{}{}
	if pageCursor != nil {
		args = append(args,
			pageCursor.LastValue,
			pageCursor.LastValue,
			pageCursor.LastID,
		)
	}
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, limit)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY %s %s LIMIT ?",
		strings.Join(selected, ", "),
		table,
		strings.Join(where, " AND "),
		orderBy,
		order,
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		row, err := ScanFunc(rows, schema)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, nil
}

func deleteRecord(db *sql.DB, table, pk string, pkVal interface{}) error {
	query := fmt.Sprintf(
		"UPDATE %s SET deleted_at = NOW() WHERE %s = ? AND deleted_at IS NULL",
		table,
		pk,
	)
	_, err := db.Exec(query, pkVal)
	return err
}

func editRecord(db *sql.DB, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	if len(cols) == 0 {
		return nil
	}

	setParts := make([]string, len(cols))
	for i, col := range cols {
		setParts[i] = fmt.Sprintf("%s = ?", col)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ? AND deleted_at IS NULL",
		table,
		strings.Join(setParts, ", "),
		pk,
	)

	vals = append(vals, pkVal)
	_, err := db.Exec(query, vals...)
	return err
}

func getRecord(db *sql.DB, id interface{}, schema map[string]string, table string, pk string, fields []string, deleted bool) (map[string]any, error) {
	selected := helper.FilterFields(fields, helper.MapKeys(schema))
	condition := "deleted_at IS NULL"
	if deleted {
		condition = "deleted_at IS NOT NULL"
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ? AND %s LIMIT 1",
		strings.Join(selected, ", "),
		table,
		pk,
		condition,
	)

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return ScanFunc(rows, schema)
	}
	return nil, sql.ErrNoRows
}

func listRecords(
	db *sql.DB,
	schema map[string]string,
	table string,
	fields []string,
	limit int,
	pageCursor *helper.PageCursor,
	orderBy, order string,
	filters []helper.Filter,
	deleted bool,
) ([]map[string]any, error) {
	selected := helper.FilterFields(fields, helper.MapKeys(schema))
	orderBy = helper.ValidateOrderBy(orderBy, helper.MapKeys(schema))
	order = helper.ValidateOrder(order)

	whereClause, args := helper.BuildWhereClause(filters)
	if deleted {
		if whereClause == "" {
			whereClause = "WHERE deleted_at IS NOT NULL"
		} else {
			whereClause += " AND deleted_at IS NOT NULL"
		}
	} else {
		if whereClause == "" {
			whereClause = "WHERE deleted_at IS NULL"
		} else {
			whereClause += " AND deleted_at IS NULL"
		}
	}

	if pageCursor != nil {
		op := ">"
		if order == "DESC" {
			op = "<"
		}
		whereClause += fmt.Sprintf(
			" AND ( %s %s ? OR ( %s = ? AND id %s ? ) )",
			orderBy, op,
			orderBy, op,
		)
		args = append(args,
			pageCursor.LastValue,
			pageCursor.LastValue,
			pageCursor.LastID,
		)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s ORDER BY %s %s LIMIT ?",
		strings.Join(selected, ", "),
		table,
		whereClause,
		orderBy,
		order,
	)
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		row, err := ScanFunc(rows, schema)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, nil
}
