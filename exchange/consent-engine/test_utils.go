package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupSQLiteTestDB creates an in-memory SQLite database for testing
func setupSQLiteTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to SQLite test database: %v", err)
	}

	// Create table with SQLite-compatible schema
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS consent_records (
		consent_id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		owner_email TEXT NOT NULL,
		app_id TEXT NOT NULL,
		status TEXT NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		pending_expires_at DATETIME,
		grant_expires_at DATETIME,
		grant_duration TEXT NOT NULL,
		fields TEXT NOT NULL,
		session_id TEXT NOT NULL,
		consent_portal_url TEXT,
		updated_by TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_consent_records_owner_id ON consent_records(owner_id);
	CREATE INDEX IF NOT EXISTS idx_consent_records_owner_email ON consent_records(owner_email);
	CREATE INDEX IF NOT EXISTS idx_consent_records_app_id ON consent_records(app_id);
	CREATE INDEX IF NOT EXISTS idx_consent_records_status ON consent_records(status);
	CREATE INDEX IF NOT EXISTS idx_consent_records_created_at ON consent_records(created_at);
	CREATE INDEX IF NOT EXISTS idx_consent_records_pending_expires_at ON consent_records(pending_expires_at);
	CREATE INDEX IF NOT EXISTS idx_consent_records_grant_expires_at ON consent_records(grant_expires_at);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Clean up test data before each test
	cleanupTestData(t, db)

	return db
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM consent_records")
	if err != nil {
		t.Logf("Warning: failed to cleanup test data: %v", err)
	}
}
