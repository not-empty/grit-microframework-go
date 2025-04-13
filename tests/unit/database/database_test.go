package database_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/not-empty/grit/app/database"
	"github.com/stretchr/testify/require"
)

func TestInit_InvalidConfig_ShouldPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic due to invalid config, got none")
		}
	}()
	_ = database.Init(database.Config{})
}

func TestLoadConfigFromEnv_Valid(t *testing.T) {
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASS", "testpass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_MAX_CONN", "10")
	t.Setenv("DB_MAX_IDLE", "5")

	cfg := database.LoadConfigFromEnv()

	require.Equal(t, "testuser", cfg.User)
	require.Equal(t, "testpass", cfg.Pass)
	require.Equal(t, "localhost", cfg.Host)
	require.Equal(t, "3306", cfg.Port)
	require.Equal(t, "testdb", cfg.Name)
	require.Equal(t, 10, cfg.MaxOpen)
	require.Equal(t, 5, cfg.MaxIdle)
}

func TestLoadConfigFromEnv_InvalidNumbers_ShouldPanic(t *testing.T) {
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASS", "pass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_MAX_CONN", "not-a-number")
	t.Setenv("DB_MAX_IDLE", "5")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic due to invalid DB_MAX_CONN")
		}
	}()

	_ = database.LoadConfigFromEnv()
}

func TestInit_WithMockedSQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing().WillReturnError(nil)

	originalOpen := database.SqlOpenFunc
	originalPing := database.DbPingFunc
	defer func() {
		database.SqlOpenFunc = originalOpen
		database.DbPingFunc = originalPing
	}()

	database.SqlOpenFunc = func(driver, dsn string) (*sql.DB, error) {
		require.Equal(t, "mysql", driver)
		return db, nil
	}

	database.DbPingFunc = func(db *sql.DB) error {
		return db.Ping()
	}

	cfg := database.Config{
		User: "user", Pass: "pass", Host: "localhost", Port: "3306",
		Name: "testdb", MaxOpen: 5, MaxIdle: 2,
	}

	conn := database.Init(cfg)
	require.NotNil(t, conn)
}

func TestInit_SQLConnectionFails_ShouldPanic(t *testing.T) {
	original := database.SqlOpenFunc
	defer func() { database.SqlOpenFunc = original }()

	database.SqlOpenFunc = func(driver, dsn string) (*sql.DB, error) {
		return nil, errors.New("sql open error")
	}

	cfg := database.Config{
		User: "user", Pass: "pass", Host: "localhost", Port: "3306",
		Name: "testdb", MaxOpen: 5, MaxIdle: 2,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic due to sql.Open failure")
		}
	}()

	_ = database.Init(cfg)
}

func TestInit_PingFails_ShouldPanic(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	originalOpen := database.SqlOpenFunc
	originalPing := database.DbPingFunc
	defer func() {
		database.SqlOpenFunc = originalOpen
		database.DbPingFunc = originalPing
	}()

	database.SqlOpenFunc = func(driver, dsn string) (*sql.DB, error) {
		return db, nil
	}

	database.DbPingFunc = func(_ *sql.DB) error {
		return errors.New("ping failure")
	}

	cfg := database.Config{
		User: "user", Pass: "pass", Host: "localhost", Port: "3306",
		Name: "testdb", MaxOpen: 5, MaxIdle: 2,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic due to Ping failure")
		}
	}()

	_ = database.Init(cfg)
}

func TestDbPingFunc_CallsPing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing().WillReturnError(nil)

	err = database.DbPingFunc(db)
	require.NoError(t, err)
}
