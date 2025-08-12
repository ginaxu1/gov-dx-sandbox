// internal/database/database.go
package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"policy-governance/internal/models"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}) // Use postgres.Open
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Database connection has been established successfully.")

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
		log.Println("Database already contains data. Skipping seeding.")
		return
	}

	log.Println("Seeding initial data...")
	providers := []models.Provider{
		{ProviderID: "drp_service", ProviderName: "DRP", IsGovtEntity: true},
		{ProviderID: "dmt_service", ProviderName: "DMT", IsGovtEntity: true},
		{ProviderID: "drc_service", ProviderName: "DRC", IsGovtEntity: true},
	}
	DB.Create(&providers)

	mappings := []models.PolicyMapping{
		{PolicyID: "policy_101", ConsumerID: "hotel_service_id", ProviderID: "drp_service", AccessTier: "Confidential", AccessBucket: "requires_consent"},
		{PolicyID: "policy_102", ConsumerID: "dmt_service_id", ProviderID: "drp_service", AccessTier: "Confidential", AccessBucket: "govt_access"},
		{PolicyID: "policy_103", ConsumerID: "citizen_app_id", ProviderID: "drp_service", AccessTier: "Confidential", AccessBucket: "access_by_consumer_id"},
		{PolicyID: "policy_104", ConsumerID: "hotel_service_id", ProviderID: "dmt_service", AccessTier: "Limited Access", AccessBucket: "requires_consent_limited"},
		{PolicyID: "policy_105", ConsumerID: "citizen_app_id", ProviderID: "dmt_service", AccessTier: "Confidential", AccessBucket: "access_by_consumer_id"},
		{PolicyID: "policy_106", ConsumerID: "dmt_service_id", ProviderID: "dmt_service", AccessTier: "Public", AccessBucket: "none"},
		{PolicyID: "policy_107", ConsumerID: "dmt_service_id", ProviderID: "drc_service", AccessTier: "Secret", AccessBucket: "platform_denial"},
		{PolicyID: "policy_108", ConsumerID: "dmt_service_id", ProviderID: "drc_service", AccessTier: "Confidential", AccessBucket: "govt_access"},
	}
	DB.Create(&mappings)
	log.Println("Initial data seeded successfully.")
}