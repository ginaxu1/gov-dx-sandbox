package repository

import (
	"context"
	"log"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"policy-governance/internal/models"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	var err error
	// Use an in-memory database for fast, isolated tests.
	testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate tables once for the entire test suite.
	err = testDB.AutoMigrate(&models.Provider{}, &models.PolicyMapping{})
	if err != nil {
		log.Fatalf("failed to migrate test database: %v", err)
	}

	// Run all tests
	exitCode := m.Run()

	// Clean up resources
	sqlDB, _ := testDB.DB()
	sqlDB.Close()

	os.Exit(exitCode)
}

func TestGetPolicy_Success(t *testing.T) {
	// Start a transaction for a clean test environment.
	// All changes made inside this block will be rolled back.
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	// Seed data directly within the test transaction.
	mappings := []models.PolicyMapping{
		{PolicyID: "policy_101", ConsumerID: "hotel_service_id", ProviderID: "drp_service", AccessTier: "Tier 2", AccessBucket: "require_consent"},
	}
	if err := tx.Create(&mappings).Error; err != nil {
		t.Fatalf("failed to seed policies: %v", err)
	}

	repo := &PolicyRepository{db: tx}
	ctx := context.Background()
	consumerID := "hotel_service_id"
	providerID := "drp_service"

	policy, err := repo.GetPolicy(ctx, consumerID, providerID)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if policy == nil {
		t.Error("expected a policy, got nil")
		return
	}

	if policy.AccessTier != "Tier 2" {
		t.Errorf("expected access tier 'Tier 2', got: %s", policy.AccessTier)
	}
	if policy.AccessBucket != "require_consent" {
		t.Errorf("expected access bucket 'require_consent', got: %s", policy.AccessBucket)
	}
}

func TestGetPolicy_NotFound(t *testing.T) {
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	repo := &PolicyRepository{db: tx}
	ctx := context.Background()
	consumerID := "non_existent_consumer"
	providerID := "drp_service"

	policy, err := repo.GetPolicy(ctx, consumerID, providerID)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if policy != nil {
		t.Errorf("expected no policy, got: %+v", policy)
	}
}