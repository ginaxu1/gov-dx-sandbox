package testhelpers

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
)

// StringPtr returns a pointer to the given string value.
// This is a convenience helper for test code that needs string pointers.
func StringPtr(s string) *string {
	return &s
}

// OwnerPtr returns a pointer to the given Owner value.
// This is a convenience helper for test code that needs Owner pointers.
func OwnerPtr(o models.Owner) *models.Owner {
	return &o
}

// SetupTestDB creates an in-memory SQLite database for testing.
// It creates the policy_metadata table with SQLite-compatible schema.
// SQLite doesn't support PostgreSQL-specific features like gen_random_uuid(), enums, jsonb.
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create table manually for SQLite compatibility
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS policy_metadata (
			id TEXT PRIMARY KEY,
			schema_id TEXT NOT NULL,
			field_name TEXT NOT NULL,
			display_name TEXT,
			description TEXT,
			source TEXT NOT NULL DEFAULT 'fallback',
			is_owner INTEGER NOT NULL DEFAULT 0,
			access_control_type TEXT NOT NULL DEFAULT 'restricted',
			allow_list TEXT NOT NULL DEFAULT '{}',
			owner TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(schema_id, field_name)
		)
	`
	if err := db.Exec(createTableSQL).Error; err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}
