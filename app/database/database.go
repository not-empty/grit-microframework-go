package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

func Init() *sql.DB {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	maxconnStr := os.Getenv("DB_MAX_CONN")
	maxidleStr := os.Getenv("DB_MAX_IDLE")
	if user == "" || pass == "" || host == "" || port == "" || dbName == "" || maxconnStr == "" || maxidleStr == "" {
		log.Fatal("Missing one or more database configuration environment variables")
	}

	maxconn, err := strconv.Atoi(maxconnStr)
	if err != nil {
		log.Fatalf("Error converting DB_MAX_CONN: %v", err)
	}
	maxidle, err := strconv.Atoi(maxidleStr)
	if err != nil {
		log.Fatalf("Error converting DB_MAX_IDLE: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	db.SetMaxOpenConns(maxconn)
	db.SetMaxIdleConns(maxidle)

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return db
}
