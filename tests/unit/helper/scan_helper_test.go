package helper

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/not-empty/grit/app/helper"
	"github.com/stretchr/testify/require"
)

type mockScannerWithColumnError struct{}

func (m *mockScannerWithColumnError) Columns() ([]string, error) {
	return nil, errors.New("simulated column error")
}

func (m *mockScannerWithColumnError) Scan(...any) error {
	return nil
}

type failingScanner struct {
	*sql.Rows
}

func (fs *failingScanner) Scan(dest ...any) error {
	return fmt.Errorf("forced scan failure")
}

func TestGenericScanToMap_ErrorGettingColumns(t *testing.T) {
	scanner := &mockScannerWithColumnError{}
	_, err := helper.GenericScanToMap(scanner, map[string]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get columns")
}

func TestGenericScanToMap_ScanFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow("test")
	mock.ExpectQuery("SELECT id").WillReturnRows(rows)

	query := "SELECT id"
	schema := map[string]string{
		"id": "string",
	}

	r, err := db.Query(query)
	require.NoError(t, err)
	require.True(t, r.Next())

	scanner := &failingScanner{r}

	result, err := helper.GenericScanToMap(scanner, schema)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "scan failed")
}

func TestGenericScanToMap_DiscardPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"unknown"}).AddRow("notused")
	mock.ExpectQuery("SELECT discard").WillReturnRows(rows)

	r, err := db.Query("SELECT discard")
	require.NoError(t, err)
	require.True(t, r.Next())

	schema := map[string]string{"id": "string"}
	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)
	require.Len(t, result, 0)
}

func TestGenericScanToMap_DefaultAssignmentNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(nil)

	mock.ExpectQuery("SELECT default").WillReturnRows(rows)

	r, err := db.Query("SELECT default")
	require.NoError(t, err)
	require.True(t, r.Next())

	schema := map[string]string{"id": "string"}
	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)
	require.Equal(t, "", result["id"])
}

func TestGenericScanToMap_TimeFormatOutput(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dt := time.Date(2023, 10, 2, 15, 4, 5, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"created_at"}).
		AddRow(dt)

	mock.ExpectQuery("SELECT date").WillReturnRows(rows)

	r, err := db.Query("SELECT date")
	require.NoError(t, err)
	require.True(t, r.Next())

	schema := map[string]string{"created_at": "*time.Time"}
	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)
	require.Equal(t, "2023-10-02 15:04:05", result["created_at"])
}

func TestGenericScanToMap_IntColumn(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"age"}).AddRow(int64(42))
	mock.ExpectQuery("SELECT age").WillReturnRows(rows)

	schema := map[string]string{
		"age": "int",
	}

	resultSet, err := db.Query("SELECT age")
	require.NoError(t, err)
	require.True(t, resultSet.Next())

	result, err := helper.GenericScanToMap(resultSet, schema)
	require.NoError(t, err)

	require.Equal(t, 42, result["age"])
}

func TestGenericScanToMap_UnsupportedTypeDefaultsToDiscard(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"active"}).AddRow(true)

	mock.ExpectQuery("SELECT \\* FROM test").WillReturnRows(rows)

	schema := map[string]string{
		"active": "bool",
	}

	r, err := db.Query("SELECT * FROM test")
	require.NoError(t, err)
	defer r.Close()

	require.True(t, r.Next())

	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)
	require.NotContains(t, result, "active")
}

func TestGenericScanToMap_NullStringValid(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"username"}).
		AddRow(sql.NullString{String: "leo", Valid: true})

	mock.ExpectQuery("SELECT \\* FROM test").WillReturnRows(rows)

	schema := map[string]string{
		"username": "string",
	}

	r, err := db.Query("SELECT * FROM test")
	require.NoError(t, err)
	defer r.Close()

	require.True(t, r.Next())

	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)

	require.Contains(t, result, "username")
	require.Equal(t, "leo", result["username"])
}

func TestGenericScanToMap_NullInt64Invalid(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"age"}).
		AddRow(sql.NullInt64{Int64: 99, Valid: false})

	mock.ExpectQuery("SELECT \\* FROM test").WillReturnRows(rows)

	schema := map[string]string{
		"age": "int",
	}

	r, err := db.Query("SELECT * FROM test")
	require.NoError(t, err)
	defer r.Close()

	require.True(t, r.Next())

	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)

	require.Contains(t, result, "age")
	require.Equal(t, 0, result["age"])
}

func TestGenericScanToMap_NullTimeInvalid(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"created_at"}).
		AddRow(sql.NullTime{Time: time.Now(), Valid: false})

	mock.ExpectQuery("SELECT \\* FROM test").WillReturnRows(rows)

	schema := map[string]string{
		"created_at": "*time.Time",
	}

	r, err := db.Query("SELECT * FROM test")
	require.NoError(t, err)
	defer r.Close()

	require.True(t, r.Next())

	result, err := helper.GenericScanToMap(r, schema)
	require.NoError(t, err)

	require.Contains(t, result, "created_at")
	require.Nil(t, result["created_at"])
}

func TestMapKeys(t *testing.T) {
	schema := map[string]string{
		"id":    "string",
		"email": "string",
		"age":   "int",
	}

	keys := helper.MapKeys(schema)
	require.ElementsMatch(t, []string{"id", "email", "age"}, keys)
	require.Len(t, keys, 3)
}
