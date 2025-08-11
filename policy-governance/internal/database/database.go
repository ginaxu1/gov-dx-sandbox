// internal/database/database.go
package database

import (
	"log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"policy-governance/internal/models"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("data/policy.sqlite"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Database connection has been established successfully")

	// Auto-migrate the database tables
	err = DB.AutoMigrate(&models.Provider{}, &models.PolicyMapping{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database tables: %v", err)
	}
	log.Println("Database synchronized. Tables created/updated.")

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&models.Provider{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data. Skipping seeding")
		return
	}

	log.Println("Seeding initial data...")
	providers := []models.Provider{
		{ProviderID: "drp_service", ProviderName: "DRP", IsGovtEntity: true},
		{ProviderID: "dmt_service", ProviderName: "DMT", IsGovtEntity: true},
	}
	DB.Create(&providers)

	mappings := []models.PolicyMapping{
		{PolicyID: "policy_101", ConsumerID: "bank_service_id", ProviderID: "drp_service", AccessTier: "Tier 2", AccessBucket: "require_consent"},
		{PolicyID: "policy_102", ConsumerID: "dmt_service_id", ProviderID: "drp_service", AccessTier: "Tier 2", AccessBucket: "govt_access"},
		{PolicyID: "policy_103", ConsumerID: "citizen_app_id", ProviderID: "drp_service", AccessTier: "Tier 2", AccessBucket: "access_by_consumer_id"},
		{PolicyID: "policy_104", ConsumerID: "bank_service_id", ProviderID: "dmt_service", AccessTier: "Tier 2", AccessBucket: "require_consent_limited"},
		{PolicyID: "policy_105", ConsumerID: "citizen_app_id", ProviderID: "dmt_service", AccessTier: "Tier 2", AccessBucket: "access_by_consumer_id"},
		{PolicyID: "policy_106", ConsumerID: "dmt_service_id", ProviderID: "dmt_service", AccessTier: "Tier 1", AccessBucket: "public"},
	}
	DB.Create(&mappings)
	log.Println("Initial data seeded successfully")
}