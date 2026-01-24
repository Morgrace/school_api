package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewDB opens the database connection and configures the pool.
// It returns the *sql.DB object so main.go can control its lifecycle.
func NewDB() (*sql.DB, error) {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")
	databaseHost := os.Getenv("DB_HOST")
	databasePort := os.Getenv("DB_PORT")

	// Pro Tip: parseTime=true is required for scanning MySQL DATETIME into Go time.Time
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify&parseTime=true",
		username, password, databaseHost, databasePort, databaseName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// ⚙️ Connection Pool Settings (Excellent choice, keeping these)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection immediately
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}
	log.Println("Connected to Database Successfully 🌐")
	return db, nil
}
