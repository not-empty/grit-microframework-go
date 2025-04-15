package repository_test

import (
	"database/sql"
	"regexp"
	"testing"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/not-empty/grit/app/repository"
	"github.com/not-empty/grit/app/helper"
	"github.com/not-empty/grit/app/repository/models"
	"github.com/stretchr/testify/require"
)

func newTestRepo(db *sql.DB) repository.RepositoryInterface[*models.Example] {
	return repository.NewRepository[*models.Example](db, func() *models.Example {
		return &models.Example{}
	})
}

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	example := &models.Example{ID: "1", Name: "John", Age: 30}
	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO example (id, name, age, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)`)).
		WithArgs(example.ID, example.Name, example.Age, nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Insert(example)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE example SET name = ? WHERE id = ? AND deleted_at IS NULL`)).
		WithArgs("Jane", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateFields("example", "id", "1", []string{"name"}, []interface{}{"Jane"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE example SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`)).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(&models.Example{ID: "1"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id = ? AND deleted_at IS NULL LIMIT 1`)).
		WithArgs("1").
		WillReturnRows(rows)

	result, err := repo.Get("1", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Equal(t, "John", result["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.ListActive(10, 0, "id", "asc", []string{"id", "name", "age"}, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id IN (?, ?) AND deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs("1", "2", 10, 0).
		WillReturnRows(rows)

	result, err := repo.BulkGet([]string{"1", "2"}, 10, 0, "id", "asc", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id = ? AND deleted_at IS NOT NULL LIMIT 1`)).
		WithArgs("1").
		WillReturnRows(rows)

	result, err := repo.GetDeleted("1", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Equal(t, "John", result["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE deleted_at IS NOT NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.ListDeleted(10, 0, "id", "asc", []string{"id", "name", "age"}, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateFieldsEmptyColumns(t *testing.T) {
    db, _, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := newTestRepo(db)

    err = repo.UpdateFields("example", "id", "1", []string{}, []interface{}{})
    require.NoError(t, err)
}

func TestGet_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id = ? AND deleted_at IS NULL LIMIT 1`)).
		WithArgs("non-existent").
		WillReturnError(sql.ErrConnDone)

	_, err = repo.Get("non-existent", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Equal(t, sql.ErrConnDone, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetModel_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id = ? AND deleted_at IS NULL LIMIT 1`)).
		WithArgs("not-found").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

	result, err := repo.Get("not-found", []string{"id", "name", "age"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListActiveWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE name = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs("John", 10, 0).
		WillReturnRows(rows)

	filters := []helper.Filter{
		{
			Field:    "name",
			Operator: "eql",
			Value:    "John",
		},
	}
	result, err := repo.ListActive(10, 0, "id", "asc", []string{"id", "name", "age"}, filters)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListDeletedWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "age"}).AddRow("1", "John", 30)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE name = ? AND deleted_at IS NOT NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs("John", 10, 0).
		WillReturnRows(rows)

	filters := []helper.Filter{
		{
			Field:    "name",
			Operator: "eql",
			Value:    "John",
		},
	}
	result, err := repo.ListDeleted(10, 0, "id", "asc", []string{"id", "name", "age"}, filters)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "John", result[0]["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListActive_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs(10, 0).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.ListActive(10, 0, "id", "asc", []string{"id", "name", "age"}, nil)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, sql.ErrConnDone, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListActive_ScanError(t *testing.T) {
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
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.ListActive(10, 0, "id", "asc", []string{"id", "name", "age"}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scan error")
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkGet_EmptyIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	result, err := repo.BulkGet([]string{}, 10, 0, "id", "asc", []string{"id", "name", "age"})
	require.NoError(t, err)
	require.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkGet_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

	expectedErr := sql.ErrConnDone
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, age FROM example WHERE id IN (?, ?) AND deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`)).
		WithArgs("1", "2", 10, 0).
		WillReturnError(expectedErr)

	result, err := repo.BulkGet([]string{"1", "2"}, 10, 0, "id", "asc", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
	require.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkGet_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := newTestRepo(db)

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
		`SELECT id, name, age FROM example WHERE id IN (?, ?) AND deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`,
	)).
		WithArgs("1", "2", 10, 0).
		WillReturnRows(rows)

	result, err := repo.BulkGet([]string{"1", "2"}, 10, 0, "id", "asc", []string{"id", "name", "age"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "scan error")
	require.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}