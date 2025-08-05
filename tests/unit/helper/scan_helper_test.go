package helper

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/not-empty/grit-microframework-go/app/helper"
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

func TestGenericScanToMap_TimeFormatOutputNullable(t *testing.T) {
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

	schema := map[string]string{"created_at": "time.Time"}
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

func TestSimpleScanRows_SingleRow_AllTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{
		"str",
		"bytes",
		"i",
		"f",
		"b",
		"t",
		"nul",
	}).AddRow(
		"hello",
		[]byte("world"),
		int64(42),
		3.1415,
		true,
		now,
		nil,
	)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	r, err := db.Query("SELECT")
	require.NoError(t, err)
	defer r.Close()

	out, err := helper.SimpleScanRows(r)
	require.NoError(t, err)
	require.Len(t, out, 1)

	row := out[0]

	require.IsType(t, "", row["str"])
	require.Equal(t, "hello", row["str"])

	require.IsType(t, "", row["bytes"])
	require.Equal(t, "world", row["bytes"])

	require.IsType(t, int64(0), row["i"])
	require.Equal(t, int64(42), row["i"])

	require.IsType(t, float64(0), row["f"])
	require.Equal(t, 3.1415, row["f"])

	require.IsType(t, bool(false), row["b"])
	require.Equal(t, true, row["b"])

	require.IsType(t, time.Time{}, row["t"])

	require.Equal(t, now, row["t"])

	require.Nil(t, row["nul"])
}

func TestSimpleScanRows_MultipleRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("alice").
		AddRow("bob")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	r, err := db.Query("SELECT")
	require.NoError(t, err)
	defer r.Close()

	out, err := helper.SimpleScanRows(r)
	require.NoError(t, err)
	require.Len(t, out, 2)

	require.Equal(t, "alice", out[0]["name"])
	require.Equal(t, "bob", out[1]["name"])
}

type errScanRS struct {
	called bool
}

func (r *errScanRS) Columns() ([]string, error) {
	return []string{"foo"}, nil
}

func (r *errScanRS) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, nil
}

func (r *errScanRS) Next() bool {
	if !r.called {
		r.called = true
		return true
	}
	return false
}

func (r *errScanRS) Scan(dest ...any) error {
	return fmt.Errorf("boom scan")
}

func (r *errScanRS) Err() error {
	return nil
}

func TestSimpleScanRows_ScanError(t *testing.T) {
	out, err := helper.SimpleScanRows(&errScanRS{})
	require.Nil(t, out)
	require.EqualError(t, err, "boom scan")
}

type errIteratorRS struct{}

func (r *errIteratorRS) Columns() ([]string, error) {
	return []string{"col"}, nil
}

func (r *errIteratorRS) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, nil
}

func (r *errIteratorRS) Next() bool {
	return false
}

func (r *errIteratorRS) Scan(dest ...any) error {
	return nil
}

func (r *errIteratorRS) Err() error {
	return errors.New("iteration failed")
}

func TestSimpleScanRows_IteratorError(t *testing.T) {
	out, err := helper.SimpleScanRows(&errIteratorRS{})
	require.Nil(t, out)
	require.EqualError(t, err, "iteration failed")
}

func TestRowsAdapter_ColumnsAndColumnTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("1", "Alice")
	mock.ExpectQuery("SELECT 1").WillReturnRows(rows)

	r, err := db.Query("SELECT 1")
	require.NoError(t, err)
	defer r.Close()

	ra := helper.NewRowsAdapter(r)

	cols, err := ra.Columns()
	require.NoError(t, err)
	require.Equal(t, []string{"id", "name"}, cols)

	cts, err := ra.ColumnTypes()
	require.NoError(t, err)
	require.Len(t, cts, 2)
	require.Equal(t, "id", cts[0].Name())
	require.Equal(t, "name", cts[1].Name())
}

func TestRowsAdapter_NextScanErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"value"}).
		AddRow(int64(42))
	mock.ExpectQuery("SELECT 2").WillReturnRows(rows)

	r, err := db.Query("SELECT 2")
	require.NoError(t, err)
	defer r.Close()

	ra := helper.NewRowsAdapter(r)

	require.True(t, ra.Next())

	var v int64
	require.NoError(t, ra.Scan(&v))
	require.Equal(t, int64(42), v)

	require.False(t, ra.Next())

	require.NoError(t, ra.Err())
}

func TestGenericScanToMap_DateColumnFormatsAsYYYYMMDD(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE t(d DATE)`)
	require.NoError(t, err)

	dt := time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)
	_, err = db.Exec(`INSERT INTO t(d) VALUES(?)`, dt.Format("2006-01-02"))
	require.NoError(t, err)

	rows, err := db.Query(`SELECT d FROM t`)
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())

	schema := map[string]string{"d": "*time.Time"}
	m, err := helper.GenericScanToMap(rows, schema)
	require.NoError(t, err)

	require.Equal(t, "2025-06-06", m["d"])
}
