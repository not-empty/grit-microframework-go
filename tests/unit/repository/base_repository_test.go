package repository_test

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/repository"
	"github.com/not-empty/grit/app/repository/models"
	"github.com/stretchr/testify/require"
)

func newTestRepo(db *sql.DB) repository.RepositoryInterface[*models.Example] {
	return repository.NewRepository[*models.Example](db, func() *models.Example {
		return &models.Example{}
	})
}

func TestAdd(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	example := &models.Example{ID: "1", Name: "John", Age: 30}
	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `example` (`id`, `name`, `age`, `created_at`, `updated_at`, `deleted_at`) VALUES (?, ?, ?, ?, ?, ?)")).
		WithArgs(example.ID, example.Name, example.Age, nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Add(example)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEdit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `example` SET `name` = ? WHERE `id` = ? AND `deleted_at` IS NULL")).
		WithArgs("Jane", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Edit("`example`", "id", "1", []string{"name"}, []interface{}{"Jane"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE `example` SET `deleted_at` = NOW() WHERE `id` = ? AND `deleted_at` IS NULL")).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&models.Example{ID: "1"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDetail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`, `name`, `age` FROM `example` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1")).
		WithArgs("1").
		WillReturnRows(rows)

	result, err := repo.Detail("1", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Equal(t, "John", result["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeadDetail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`, `name`, `age` FROM `example` WHERE `id` = ? AND `deleted_at` IS NOT NULL LIMIT 1")).
		WithArgs("1").
		WillReturnRows(rows)

	result, err := repo.DeadDetail("1", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Equal(t, "John", result["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEditEmptyColumns(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	err = repo.Edit("example", "id", "1", []string{}, []interface{}{})
	require.NoError(t, err)
}

func TestDetail_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`, `name`, `age` FROM `example` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1")).
		WithArgs("non-existent").
		WillReturnError(sql.ErrConnDone)

	_, err = repo.Detail("non-existent", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Equal(t, sql.ErrConnDone, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDetaill_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`, `name`, `age` FROM `example` WHERE `id` = ? AND `deleted_at` IS NULL LIMIT 1")).
		WithArgs("not-found").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

	result, err := repo.Detail("not-found", []string{"id", "name", "age"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs(10).
		WillReturnRows(rows)

	result, err := repo.List(10, nil, "id", "asc", []string{"id", "name", "age"}, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeadList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NOT NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs(10).
		WillReturnRows(rows)

	result, err := repo.DeadList(10, nil, "id", "asc", []string{"id", "name", "age"}, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `name` = ? AND `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs("John", 10).
		WillReturnRows(rows)

	filters := []helper.Filter{{Field: "name", Operator: "eql", Value: "John"}}
	result, err := repo.List(10, nil, "id", "asc", []string{"id", "name", "age"}, filters)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeadListWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `name` = ? AND `deleted_at` IS NOT NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs("John", 10).
		WillReturnRows(rows)

	filters := []helper.Filter{{Field: "name", Operator: "eql", Value: "John"}}
	result, err := repo.DeadList(10, nil, "id", "asc", []string{"id", "name", "age"}, filters)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs(10).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.List(10, nil, "id", "asc", []string{"id", "name", "age"}, nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, sql.ErrConnDone, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	originalScan := repository.ScanFunc
	repository.ScanFunc = func(scanner interface {
		Columns() ([]string, error)
		Scan(...any) error
	}, schema map[string]string) (map[string]any, error) {
		return nil, fmt.Errorf("scan error")
	}
	defer func() { repository.ScanFunc = originalScan }()

	repo := newTestRepo(db)
	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs(10).
		WillReturnRows(rows)

	result, err := repo.List(10, nil, "id", "asc", []string{"id", "name", "age"}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scan error")
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL AND `id` IN (?, ?) ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs("1", "2", 10).
		WillReturnRows(rows)

	result, err := repo.Bulk([]string{"1", "2"}, 10, nil, "id", "asc", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk_EmptyIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	result, err := repo.Bulk([]string{}, 10, nil, "id", "asc", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	expectedErr := sql.ErrConnDone
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL AND `id` IN (?, ?) ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs("1", "2", 10).
		WillReturnError(expectedErr)

	result, err := repo.Bulk([]string{"1", "2"}, 10, nil, "id", "asc", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	originalScan := repository.ScanFunc
	repository.ScanFunc = func(scanner interface {
		Columns() ([]string, error)
		Scan(...interface{}) error
	}, schema map[string]string) (map[string]any, error) {
		return nil, fmt.Errorf("scan error")
	}
	defer func() { repository.ScanFunc = originalScan }()

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL AND `id` IN (?, ?) ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs("1", "2", 10).
		WillReturnRows(rows)

	repo := newTestRepo(db)

	result, err := repo.Bulk([]string{"1", "2"}, 10, nil, "id", "asc", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "scan error")
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_WithCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	cursor := &helper.PageCursor{
		LastID:    "2",
		LastValue: "42",
	}

	query := regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` " +
			"WHERE `deleted_at` IS NULL AND ( `id` < ? OR ( `id` = ? AND `id` < ? ) ) " +
			"ORDER BY `id` DESC LIMIT ?",
	)

	mock.ExpectQuery(query).
		WithArgs(cursor.LastValue, cursor.LastValue, cursor.LastID, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow("3", "Alice", 25),
		)

	list, err := repo.List(10, cursor, "id", "desc", []string{"id", "name", "age"}, nil)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Alice", list[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk_WithCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	cursor := &helper.PageCursor{
		LastID:    "100",
		LastValue: "20",
	}
	ids := []string{"1", "2"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL "+
			"AND (`id` > ? OR (`id` = ? AND `id` > ?)) "+
			"AND `id` IN (?, ?) ORDER BY `id` ASC LIMIT ?",
	)).
		WithArgs(
			cursor.LastValue,
			cursor.LastValue,
			cursor.LastID,
			ids[0],
			ids[1],
			5,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow("2", "Bob", 28),
		)

	list, err := repo.Bulk(ids, 5, cursor, "id", "ASC", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Bob", list[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulk_WithCursorDesc(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	cursor := &helper.PageCursor{
		LastID:    "100",
		LastValue: "20",
	}
	ids := []string{"1", "2"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL "+
			"AND (`id` < ? OR (`id` = ? AND `id` < ?)) "+
			"AND `id` IN (?, ?) ORDER BY `id` DESC LIMIT ?",
	)).
		WithArgs(
			cursor.LastValue,
			cursor.LastValue,
			cursor.LastID,
			ids[0],
			ids[1],
			5,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow("2", "Bob", 28),
		)

	list, err := repo.Bulk(ids, 5, cursor, "id", "DESC", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Bob", list[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListOne_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	fields := []string{"id", "name", "age"}

	rows := sqlmock.NewRows(fields).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` ASC LIMIT ?",
	)).WithArgs(1).WillReturnRows(rows)

	result, err := repo.ListOne("id", "ASC", fields, nil)
	require.NoError(t, err)
	require.Equal(t, "John", result["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListOne_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	fields := []string{"id", "name", "age"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` ASC LIMIT ?",
	)).WithArgs(1).WillReturnRows(sqlmock.NewRows(fields))

	result, err := repo.ListOne("id", "ASC", fields, nil)
	require.NoError(t, err)
	require.Empty(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListOne_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	fields := []string{"id", "name", "age"}

	expErr := sql.ErrConnDone
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT `id`, `name`, `age` FROM `example` WHERE `deleted_at` IS NULL ORDER BY `id` ASC LIMIT ?",
	)).WithArgs(1).WillReturnError(expErr)

	result, err := repo.ListOne("id", "ASC", fields, nil)
	require.Error(t, err)
	require.Equal(t, expErr, err)
	require.Empty(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Raw_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rawQuery := "SELECT id, name FROM `example` WHERE name = :name"
	params := map[string]any{"name": "Alice"}

	convertedSQL, args := helper.PrepareRawQuery(rawQuery, params)

	mock.ExpectQuery(regexp.QuoteMeta(convertedSQL)).
		WithArgs(args[0]).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "Alice"),
		)

	results, err := repo.Raw(rawQuery, params)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Alice", results[0]["name"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Raw_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rawQuery := "SELECT id FROM `example` WHERE foo = :foo"
	params := map[string]any{"foo": "bar"}

	convertedSQL, args := helper.PrepareRawQuery(rawQuery, params)

	mock.ExpectQuery(regexp.QuoteMeta(convertedSQL)).
		WithArgs(args[0]).
		WillReturnError(fmt.Errorf("db exploded"))

	results, err := repo.Raw(rawQuery, params)
	require.Error(t, err)
	require.Nil(t, results)
	require.Contains(t, err.Error(), "db exploded")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkAdd_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	nowTime := time.Now().Truncate(time.Second)
	now1 := helper.JSONTime(nowTime)
	e1 := &models.Example{
		ID:        "1",
		Name:      "Alice",
		Age:       25,
		LastLogin: &now1,
	}

	e2 := &models.Example{
		ID:        "2",
		Name:      "Bob",
		Age:       10,
		LastLogin: nil,
	}

	sqlLiteral :=
		"INSERT INTO `example` (`id`, `name`, `age`, `last_login`, `created_at`, `updated_at`, `deleted_at`) VALUES " +
			"(?, ?, ?, ?, ?, ?, ?), (?, ?, ?, DEFAULT, ?, ?, ?)"

	pattern := regexp.QuoteMeta(sqlLiteral)

	mock.ExpectExec(pattern).
		WithArgs(
			"1", "Alice", 25, now1, nil, nil, nil,
			"2", "Bob", 10, nil, nil, nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err = repo.BulkAdd([]*models.Example{e1, e2})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkAdd_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	e := &models.Example{ID: "x", Name: "Y", Age: 10, LastLogin: nil}

	errorSQL := "INSERT INTO `example` (`id`, `name`, `age`, `last_login`, `created_at`, `updated_at`, `deleted_at`) VALUES " +
		"(?, ?, ?, DEFAULT, ?, ?, ?)"

	pattern := regexp.QuoteMeta(errorSQL)

	mock.ExpectExec(pattern).
		WithArgs("x", "Y", 10, nil, nil, nil).
		WillReturnError(fmt.Errorf("insert failed"))

	err = repo.BulkAdd([]*models.Example{e})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insert failed")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkAdd_EmptySlicePanics(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)
	require.Panics(t, func() {
		_ = repo.BulkAdd([]*models.Example{})
	})
}

func TestBuildRowTokens_ExampleModel(t *testing.T) {
	allCols := []string{"id", "name", "age", "last_login", "created_at", "updated_at", "deleted_at"}
	defaultCols := []string{"last_login"}

	nowTime := time.Now().Truncate(time.Second)
	now := helper.JSONTime(nowTime)
	vals1 := []interface{}{"A", "B", 5, &now, nil, nil, nil}
	rowSQL1, args1 := helper.BuildRowTokens(allCols, vals1, defaultCols)
	require.Equal(t, "(?, ?, ?, ?, ?, ?, ?)", rowSQL1)
	require.Equal(t, []interface{}{"A", "B", 5, &now, nil, nil, nil}, args1)

	vals2 := []interface{}{"X", "Y", 0, nil, nil, nil, nil}
	rowSQL2, args2 := helper.BuildRowTokens(allCols, vals2, defaultCols)
	require.Equal(t, "(?, ?, ?, DEFAULT, ?, ?, ?)", rowSQL2)
	require.Equal(t, []interface{}{"X", "Y", 0, nil, nil, nil}, args2)
}
