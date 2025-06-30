// ABOUTME: Test helper functions for database setup and teardown
// ABOUTME: Provides common test utilities for all test suites
package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *App {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTable := `
	CREATE TABLE names (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`

	if _, err := db.Exec(createTable); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return &App{db: db}
}

func teardownTestDB(app *App) {
	app.db.Close()
}