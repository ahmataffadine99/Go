package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func InitDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	var pingErr error
	for i := 1; i <= 10; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("waiting for database connection (%d/10)...: %v", i, pingErr)
		time.Sleep(1 * time.Second)
	}

	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("schema migration line note: %v", err)
		}
	}

	return nil
}

func RebindQuery(driverName, query string) string {
	if driverName != "postgres" {
		return query
	}
	var b strings.Builder
	paramIdx := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			b.WriteString(fmt.Sprintf("$%d", paramIdx))
			paramIdx++
		} else {
			b.WriteByte(query[i])
		}
	}
	return b.String()
}
