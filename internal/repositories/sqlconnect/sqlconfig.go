package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb() (*sql.DB, error) {
	

	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")
	databaseHost := os.Getenv("DB_HOST")
	databasePort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify&parseTime=true", username, password, databaseHost, databasePort, databaseName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	// Actually test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Connected to avnadmin MYSQL")
	return db, nil
}
