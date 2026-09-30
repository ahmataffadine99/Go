package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func InitDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
