package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
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

func LoadConfigFromEnv() Config {
	maxconn, err1 := strconv.Atoi(os.Getenv("DB_MAX_CONN"))
	maxidle, err2 := strconv.Atoi(os.Getenv("DB_MAX_IDLE"))

	if err1 != nil || err2 != nil {
		panic(fmt.Sprintf("Invalid DB_MAX_CONN or DB_MAX_IDLE: %v %v", err1, err2))
	}

	return Config{
		User:    os.Getenv("DB_USER"),
		Pass:    os.Getenv("DB_PASS"),
		Host:    os.Getenv("DB_HOST"),
		Port:    os.Getenv("DB_PORT"),
		Name:    os.Getenv("DB_NAME"),
		MaxOpen: maxconn,
		MaxIdle: maxidle,
	}
}

func Init(cfg Config) *sql.DB {
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
