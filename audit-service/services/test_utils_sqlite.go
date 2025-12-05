package services

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupSQLiteTestDB creates an in-memory SQLite database for testing.
// It automatically migrates the necessary models and cleans up data before each test.
// Note: SQLite doesn't support PostgreSQL-specific features like uuid, jsonb, etc.
// We use simplified table structures for testing.
func SetupSQLiteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create simplified tables for SQLite (without PostgreSQL-specific features)
	// ManagementEvent table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS management_events (
			id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL CHECK(event_type IN ('CREATE', 'UPDATE', 'DELETE')),
			status TEXT NOT NULL CHECK(status IN ('success', 'failure')),
			timestamp DATETIME NOT NULL,
			actor_type TEXT NOT NULL CHECK(actor_type IN ('USER', 'SERVICE')),
			actor_id TEXT,
			actor_role TEXT CHECK(actor_role IN ('MEMBER', 'ADMIN')),
			target_resource TEXT NOT NULL CHECK(target_resource IN ('MEMBERS', 'SCHEMAS', 'SCHEMA-SUBMISSIONS', 'APPLICATIONS', 'APPLICATION-SUBMISSIONS', 'POLICY-METADATA')),
			target_resource_id TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create management_events table: %v", err)
	}

	// DataExchangeEvent table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data_exchange_events (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			status TEXT NOT NULL CHECK(status IN ('success', 'failure')),
			application_id TEXT NOT NULL,
			schema_id TEXT NOT NULL,
			requested_data TEXT NOT NULL,
			on_behalf_of_owner_id TEXT,
			consumer_id TEXT,
			provider_id TEXT,
			additional_info TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create data_exchange_events table: %v", err)
	}

	// Clean up test data before each test
	t.Cleanup(func() {
		CleanupTestData(t, db)
	})

	return db
}
