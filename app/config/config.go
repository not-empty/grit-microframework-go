package config

import (
	"fmt"
	"os"
	"strconv"
)

var AppConfig *Config

type Config struct {
	AppEnv    string
	AppLog    bool
	AppNoAuth bool
	AppPort   string

	DBDriver  string
	DBHost    string
	DBMaxConn int
	DBMaxIdle int
	DBName    string
	DBPass    string
	DBPort    string
	DBUser    string

	DBHostTest string
	DBNameTest string
	DBPassTest string
	DBPortTest string
	DBUserTest string

	JwtAppSecret string
	JwtExpire    int64
	JwtRenew     int64
}

func LoadConfig() *Config {
	c := &Config{
		AppEnv:    GetEnvStr("APP_ENV", "local"),
		AppLog:    GetEnvBool("APP_LOG", true),
		AppNoAuth: GetEnvBool("APP_NO_AUTH", false),
		AppPort:   GetEnvStr("APP_PORT", "8001"),

		DBDriver:  GetEnvStr("DB_DRIVER", "mysql"),
		DBHost:    GetEnvStr("DB_HOST", "grit-mysql"),
		DBMaxConn: GetEnvInt("DB_MAX_CONN", 100),
		DBMaxIdle: GetEnvInt("DB_MAX_IDLE", 100),
		DBName:    GetEnvStr("DB_NAME", "grit"),
		DBPass:    GetEnvStr("DB_PASS", ""),
		DBPort:    GetEnvStr("DB_PORT", "3306"),
		DBUser:    GetEnvStr("DB_USER", "root"),

		DBHostTest: GetEnvStr("DB_HOST_TEST", "grit-mysql"),
		DBNameTest: GetEnvStr("DB_NAME_TEST", "grit"),
		DBPassTest: GetEnvStr("DB_PASS_TEST", ""),
		DBPortTest: GetEnvStr("DB_PORT_TEST", "3306"),
		DBUserTest: GetEnvStr("DB_USER_TEST", "root"),

		JwtAppSecret: GetEnvStr("JWT_APP_SECRET", "secret"),
		JwtExpire:    GetEnvInt64("JWT_EXPIRE", 9000),
		JwtRenew:     GetEnvInt64("JWT_RENEW", 6000),
	}

	AppConfig = c
	return c
}

func GetEnvStr(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func GetEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	if val == "true" || val == "1" {
		return true
	}
	return false
}

func GetEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var result int
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		return defaultVal
	}
	return result
}

func GetEnvInt64(key string, defaultVal int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}
	return result
}
