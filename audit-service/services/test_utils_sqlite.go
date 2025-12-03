package services

import (
	"testing"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupSQLiteTestDB creates an in-memory SQLite database for testing
func SetupSQLiteTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.ManagementEvent{},
		&models.DataExchangeEvent{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Clean up test data before each test
	CleanupTestData(t, db)

	return db
}
