package dbserver

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// dbPath specifies the location of the database file
var dbPath = "../internal/db/literary_lions.db"

func OpenDB() (*sql.DB, error) {
	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Read the SQL commands from the createTables.sql file
	createSQL, err := os.ReadFile("../internal/db/createTables.sql")
	if err != nil {
		log.Fatal(err)
	}
	// Execute the SQL commands to create the tables
	_, err = db.Exec(string(createSQL))
	if err != nil {
		log.Fatalf("failed to execute SQL commands: %v", err)
	}

	return db, nil
}

func CheckDB() (bool, error) {
	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if err == nil {
		return true, nil // Database file exists
	} else if os.IsNotExist(err) {
		return false, nil // Database file does not exist
	} else {
		return false, fmt.Errorf("failed to check database: %w", err) // An error occurred while checking
	}
}

func CreateDB() error {
	// Create a new database file
	_, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	return nil
}
