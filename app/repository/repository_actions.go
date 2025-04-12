package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/not-empty/grit/app/helper"
)

func insertModel(db *sql.DB, m BaseModel) error {
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

func updateModelFields(db *sql.DB, table, pk string, pkVal interface{}, cols []string, vals []interface{}) error {
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

func softDeleteModel(db *sql.DB, table, pk string, pkVal interface{}) error {
	query := fmt.Sprintf(
		"UPDATE %s SET deleted_at = NOW() WHERE %s = ? AND deleted_at IS NULL",
		table,
		pk,
	)
	_, err := db.Exec(query, pkVal)
	return err
}

func getModel[T BaseModel](db *sql.DB, id interface{}, m T, fields []string, deleted bool) error {
	allCols := m.Columns()
	selected := helper.FilterFields(fields, allCols)
	condition := "deleted_at IS NULL"
	if deleted {
		condition = "deleted_at IS NOT NULL"
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ? AND %s LIMIT 1",
		strings.Join(selected, ", "),
		m.TableName(),
		m.PrimaryKey(),
		condition,
	)

	rows, err := db.Query(query, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return helper.GenericScanFrom(rows, m)
	}
	return sql.ErrNoRows
}

func listModels[T BaseModel](
	db *sql.DB,
	factory func() T,
	table string,
	allCols []string,
	fields []string,
	limit, offset int,
	orderBy, order string,
	filters []helper.Filter,
	deleted bool,
) ([]T, error) {
	selected := helper.FilterFields(fields, allCols)
	orderBy = helper.ValidateOrderBy(orderBy, allCols)
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

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s ORDER BY %s %s LIMIT ? OFFSET ?",
		strings.Join(selected, ", "),
		table,
		whereClause,
		orderBy,
		order,
	)

	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []T
	for rows.Next() {
		item := factory()
		if err := helper.GenericScanFrom(rows, item); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

func bulkGetModels[T BaseModel](
	db *sql.DB,
	factory func() T,
	table string,
	allCols []string,
	fields []string,
	ids []string,
	limit, offset int,
	orderBy, order string,
) ([]T, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	selected := helper.FilterFields(fields, allCols)
	orderBy = helper.ValidateOrderBy(orderBy, allCols)
	order = helper.ValidateOrder(order)

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	args = append(args, limit, offset)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s IN (%s) AND deleted_at IS NULL ORDER BY %s %s LIMIT ? OFFSET ?",
		strings.Join(selected, ", "),
		table,
		factory().PrimaryKey(),
		strings.Join(placeholders, ", "),
		orderBy,
		order,
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []T
	for rows.Next() {
		item := factory()
		if err := helper.GenericScanFrom(rows, item); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}