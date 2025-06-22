package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(dataSourceName string) (*DB, error) {
	if dataSourceName == "" {
		dataSourceName = "claude-code-pull-worker.db"
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Note: Migrations are now handled by sqlite3def via make migrate command
	// Run `make migrate` to apply schema changes

	return &DB{db}, nil
}