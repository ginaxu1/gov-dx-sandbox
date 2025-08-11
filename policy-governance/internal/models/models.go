// internal/models/models.go
package models

// Provider struct for the 'providers' table
type Provider struct {
	ProviderID   string `gorm:"primaryKey;type:varchar(255)"`
	ProviderName string `gorm:"type:varchar(255);not null"`
	IsGovtEntity bool   `gorm:"not null;default:false"`
}

// PolicyMapping struct for the 'consumer_provider_mappings' table
type PolicyMapping struct {
	PolicyID     string `gorm:"primaryKey;type:varchar(255)"`
	ConsumerID   string `gorm:"type:varchar(255);not null"`
	ProviderID   string `gorm:"type:varchar(255);not null"`
	AccessTier   string `gorm:"type:varchar(255);not null"`
	AccessBucket string `gorm:"type:varchar(255);not null"`
}