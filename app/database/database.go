package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/not-empty/grit/app/config"
)

type DatabaseConfig struct {
	User    string
	Pass    string
	Host    string
	Port    string
	Name    string
	MaxOpen int
	MaxIdle int
}

var SqlOpenFunc = sql.Open
var DbPingFunc = func(db *sql.DB) error {
	return db.Ping()
}

func LoadDatabaseConfig() DatabaseConfig {
	appEnv := config.AppConfig.AppEnv
	isTest := appEnv == "test"

	user := config.AppConfig.DBUser
	pass := config.AppConfig.DBPass
	host := config.AppConfig.DBHost
	port := config.AppConfig.DBPort
	name := config.AppConfig.DBName

	if isTest {

		user = config.AppConfig.DBUserTest
		pass = config.AppConfig.DBPassTest
		host = config.AppConfig.DBHostTest
		port = config.AppConfig.DBPortTest
		name = config.AppConfig.DBNameTest
	}

	return DatabaseConfig{
		User:    user,
		Pass:    pass,
		Host:    host,
		Port:    port,
		Name:    name,
		MaxOpen: config.AppConfig.DBMaxConn,
		MaxIdle: config.AppConfig.DBMaxIdle,
	}
}

func Init(cfg DatabaseConfig) *sql.DB {
	if cfg.User == "" || cfg.Pass == "" || cfg.Host == "" || cfg.Port == "" || cfg.Name == "" || cfg.MaxOpen <= 0 || cfg.MaxIdle < 0 {
		panic("Missing or invalid database configuration")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)
	db, err := SqlOpenFunc("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)

	if err := DbPingFunc(db); err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}

	return db
}
