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
	// AccessTier is LIFe-compliant, follows the Data Sharing Policy's Classification levels
	AccessTier   string `gorm:"type:varchar(255);not null"` // e.g., "Public", "Limited Access", "Confidential", "Secret"
	// AccessBucket can be used for sub-classifications or specific consent types within a tier
	AccessBucket string `gorm:"type:varchar(255);not null"` // e.g., "none", "requires_consent", "govt_access", "access_by_consumer_id"
}