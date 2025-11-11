package services

import (
	"testing"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate all models
	// Note: SQLite doesn't support JSONB, so we use JSON type instead
	err = db.AutoMigrate(
		&models.Schema{},
		&models.Application{},
		&models.PDPJob{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// For SQLite, we need to handle JSONB differently
	// Convert JSONB columns to TEXT for SQLite compatibility
	if db.Dialector.Name() == "sqlite" {
		// For SQLite, JSONB is stored as TEXT
		// GORM should handle this automatically, but we ensure it works
	}

	return db
}

// mockPDPService is a mock implementation of PDPClient for testing
// Ensure it implements PDPClient interface
var _ PDPClient = (*mockPDPService)(nil)

type mockPDPService struct {
	createPolicyMetadataFunc func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error)
	updateAllowListFunc      func(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error)
}

func (m *mockPDPService) CreatePolicyMetadata(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
	if m.createPolicyMetadataFunc != nil {
		return m.createPolicyMetadataFunc(schemaID, sdl)
	}
	return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
}

func (m *mockPDPService) UpdateAllowList(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
	if m.updateAllowListFunc != nil {
		return m.updateAllowListFunc(request)
	}
	return &models.AllowListUpdateResponse{Records: []models.AllowListUpdateResponseRecord{}}, nil
}

func (m *mockPDPService) HealthCheck() error {
	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// mockAlertNotifier is a test implementation of AlertNotifier
type mockAlertNotifier struct {
	alerts []alertCall
}

type alertCall struct {
	severity string
	message  string
	details  map[string]interface{}
}

func (m *mockAlertNotifier) SendAlert(severity string, message string, details map[string]interface{}) error {
	m.alerts = append(m.alerts, alertCall{
		severity: severity,
		message:  message,
		details:  details,
	})
	return nil
}

func (m *mockAlertNotifier) reset() {
	m.alerts = []alertCall{}
}
