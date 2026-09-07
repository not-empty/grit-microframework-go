package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/not-empty/grit-microframework-go/app/helper"
)

var ScanFunc = helper.GenericScanToMap

func addRecord(ctx context.Context, db *sql.DB, m BaseModel) error {
	allCols := m.Columns()
	allVals := m.Values()

	defaultCols := m.HasDefaultValue()

	finalCols, finalVals := helper.FilterOutDefaulted(allCols, allVals, defaultCols)

	placeholders := make([]string, len(finalCols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.TableName(),
		strings.Join(helper.EscapeMysqlFields(finalCols), ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.ExecContext(ctx, query, finalVals...)
	return err
}

func bulkRecords(
	ctx context.Context,
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

	selected := helper.EscapeMysqlFields(helper.FilterFields(fields, helper.MapKeys(schema)))
	orderByEsc := helper.EscapeMysqlField(helper.ValidateOrderBy(orderBy, helper.MapKeys(schema)))
	order = helper.ValidateOrder(order)

	where := []string{"`deleted_at` IS NULL"}

	if pageCursor != nil {
		op := ">"
		if order == "DESC" {
			op = "<"
		}
		where = append(where, fmt.Sprintf(
			"(%s %s ? OR (%s = ? AND `%s` %s ?))",
			orderByEsc, op,
			orderByEsc, pk, op,
		))
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}
	where = append(where,
		fmt.Sprintf("`%s` IN (%s)", pk, strings.Join(placeholders, ", ")),
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

	orderExpr := fmt.Sprintf("%s %s", orderByEsc, order)
	if orderByEsc != "`id`" {
		orderExpr = fmt.Sprintf("%s %s, `id` %s", orderByEsc, order, order)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT ?",
		strings.Join(selected, ", "),
		table,
		strings.Join(where, " AND "),
		orderExpr,
	)

	rows, err := db.QueryContext(ctx, query, args...)
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

func bulkAddRecords(ctx context.Context, db *sql.DB, m []BaseModel) error {
	first := m[0]
	table := first.TableName()
	allCols := first.Columns()
	defaultCols := first.HasDefaultValue()

	var (
		rowsSQL []string
		args    []interface{}
	)

	for _, m := range m {
		vals := m.Values()
		rowSQL, rowArgs := helper.BuildRowTokens(allCols, vals, defaultCols)
		rowsSQL = append(rowsSQL, rowSQL)
		args = append(args, rowArgs...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(helper.EscapeMysqlFields(allCols), ", "),
		strings.Join(rowsSQL, ", "),
	)

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func deleteRecord(ctx context.Context, db *sql.DB, table, pk string, pkVal interface{}) error {
	query := fmt.Sprintf(
		"UPDATE %s SET `deleted_at` = NOW() WHERE `%s` = ? AND `deleted_at` IS NULL",
		table,
		pk,
	)
	_, err := db.ExecContext(ctx, query, pkVal)
	return err
}

func editRecord(ctx context.Context, db *sql.DB, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
	if len(cols) == 0 {
		return nil
	}

	setParts := make([]string, len(cols))
	for i, col := range cols {
		setParts[i] = fmt.Sprintf("`%s` = ?", col)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE `%s` = ? AND `deleted_at` IS NULL",
		table,
		strings.Join(setParts, ", "),
		pk,
	)

	vals = append(vals, pkVal)
	_, err := db.ExecContext(ctx, query, vals...)
	return err
}

func getRecord(ctx context.Context, db *sql.DB, id interface{}, schema map[string]string, table string, pk string, fields []string, deleted bool) (map[string]any, error) {
	selected := helper.EscapeMysqlFields(helper.FilterFields(fields, helper.MapKeys(schema)))
	condition := "`deleted_at` IS NULL"
	if deleted {
		condition = "`deleted_at` IS NOT NULL"
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE `%s` = ? AND %s LIMIT 1",
		strings.Join(selected, ", "),
		table,
		pk,
		condition,
	)

	rows, err := db.QueryContext(ctx, query, id)
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
	ctx context.Context,
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
	selected := helper.EscapeMysqlFields(helper.FilterFields(fields, helper.MapKeys(schema)))
	orderByEsc := helper.EscapeMysqlField(helper.ValidateOrderBy(orderBy, helper.MapKeys(schema)))
	order = helper.ValidateOrder(order)

	whereClause, args := helper.BuildWhereClause(filters)
	if deleted {
		if whereClause == "" {
			whereClause = "WHERE `deleted_at` IS NOT NULL"
		} else {
			whereClause += " AND `deleted_at` IS NOT NULL"
		}
	} else {
		if whereClause == "" {
			whereClause = "WHERE `deleted_at` IS NULL"
		} else {
			whereClause += " AND `deleted_at` IS NULL"
		}
	}

	if pageCursor != nil {
		op := ">"
		if order == "DESC" {
			op = "<"
		}
		whereClause += fmt.Sprintf(
			" AND ( %s %s ? OR ( %s = ? AND `id` %s ? ) )",
			orderByEsc, op,
			orderByEsc,
			op,
		)
		args = append(args,
			pageCursor.LastValue,
			pageCursor.LastValue,
			pageCursor.LastID,
		)
	}
	orderExpr := fmt.Sprintf("%s %s", orderByEsc, order)
	if orderByEsc != "`id`" {
		orderExpr = fmt.Sprintf("%s %s, `id` %s", orderByEsc, order, order)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s ORDER BY %s LIMIT ?",
		strings.Join(selected, ", "),
		table,
		whereClause,
		orderExpr,
	)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
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

func rawRecords(ctx context.Context, db *sql.DB, _ map[string]string, sqlText string, args ...interface{}) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return helper.SimpleScanRows(helper.NewRowsAdapter(rows))
}

func undeleteRecord(ctx context.Context, db *sql.DB, table, pk string, pkVal interface{}) error {
	query := fmt.Sprintf(
		"UPDATE %s SET `deleted_at` = NULL WHERE `%s` = ? AND `deleted_at` IS NOT NULL",
		table,
		pk,
	)
	_, err := db.ExecContext(ctx, query, pkVal)
	return err
}
