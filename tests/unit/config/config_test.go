package config

import (
	"testing"

	"github.com/not-empty/grit/app/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_LOG", "true")
	t.Setenv("APP_NO_AUTH", "false")
	t.Setenv("APP_PORT", "9000")

	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_MAX_CONN", "50")
	t.Setenv("DB_MAX_IDLE", "25")
	t.Setenv("DB_NAME", "prod_db")
	t.Setenv("DB_PASS", "password")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "admin")

	t.Setenv("DB_HOST_TEST", "localhost_test")
	t.Setenv("DB_NAME_TEST", "test_db")
	t.Setenv("DB_PASS_TEST", "test_pass")
	t.Setenv("DB_PORT_TEST", "15432")
	t.Setenv("DB_USER_TEST", "test_admin")

	t.Setenv("JWT_APP_SECRET", "supersecret")
	t.Setenv("JWT_EXPIRE", "7200")
	t.Setenv("JWT_RENEW", "3600")

	cfg := config.LoadConfig()

	require.Equal(t, "production", cfg.AppEnv)
	require.True(t, cfg.AppLog)
	require.False(t, cfg.AppNoAuth)
	require.Equal(t, "9000", cfg.AppPort)

	require.Equal(t, "postgres", cfg.DBDriver)
	require.Equal(t, "localhost", cfg.DBHost)
	require.Equal(t, 50, cfg.DBMaxConn)
	require.Equal(t, 25, cfg.DBMaxIdle)
	require.Equal(t, "prod_db", cfg.DBName)
	require.Equal(t, "password", cfg.DBPass)
	require.Equal(t, "5432", cfg.DBPort)
	require.Equal(t, "admin", cfg.DBUser)

	require.Equal(t, "localhost_test", cfg.DBHostTest)
	require.Equal(t, "test_db", cfg.DBNameTest)
	require.Equal(t, "test_pass", cfg.DBPassTest)
	require.Equal(t, "15432", cfg.DBPortTest)
	require.Equal(t, "test_admin", cfg.DBUserTest)

	require.Equal(t, "supersecret", cfg.JwtAppSecret)
	require.Equal(t, int64(7200), cfg.JwtExpire)
	require.Equal(t, int64(3600), cfg.JwtRenew)
}

func TestGetEnvStr(t *testing.T) {
	t.Setenv("TEST_ENV_VAR", "value123")
	require.Equal(t, "value123", config.GetEnvStr("TEST_ENV_VAR", "default"))
	require.Equal(t, "default", config.GetEnvStr("NON_EXISTENT_VAR", "default"))
}

func TestGetEnvBool(t *testing.T) {
	t.Setenv("BOOL_TRUE", "true")
	t.Setenv("BOOL_ONE", "1")
	t.Setenv("BOOL_FALSE", "false")

	require.True(t, config.GetEnvBool("BOOL_TRUE", false))
	require.True(t, config.GetEnvBool("BOOL_ONE", false))
	require.False(t, config.GetEnvBool("BOOL_FALSE", true))
	require.True(t, config.GetEnvBool("NON_EXISTENT_BOOL", true))
	require.False(t, config.GetEnvBool("NON_EXISTENT_BOOL_FALSE", false))
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("INT_VALID", "42")
	t.Setenv("INT_INVALID", "notanint")

	require.Equal(t, 42, config.GetEnvInt("INT_VALID", 10))
	require.Equal(t, 10, config.GetEnvInt("INT_MISSING", 10))
	require.Equal(t, 20, config.GetEnvInt("INT_INVALID", 20))
}

func TestGetEnvInt64(t *testing.T) {
	t.Setenv("INT_VALID", "42")
	t.Setenv("INT_INVALID", "notanint")

	require.Equal(t, int64(42), config.GetEnvInt64("INT_VALID", 10))
	require.Equal(t, int64(10), config.GetEnvInt64("INT_MISSING", 10))
	require.Equal(t, int64(20), config.GetEnvInt64("INT_INVALID", 20))
}
