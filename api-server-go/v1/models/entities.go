package models

// Entity represents the normalized entity table
type Entity struct {
	EntityID    string `gorm:"primarykey;column:entity_id" json:"entityId"`
	Name        string `gorm:"column:name;not null" json:"name"`
	EntityType  string `gorm:"column:entity_type;not null" json:"entityType"`
	Email       string `gorm:"column:email;not null" json:"email"`
	PhoneNumber string `gorm:"column:phone_number;not null" json:"phoneNumber"`
	BaseModel
}

// TableName sets the table name for GORM
func (Entity) TableName() string {
	return "entities"
}

// Provider represents the providers table
type Provider struct {
	ProviderID string `gorm:"primarykey;column:provider_id" json:"providerId"`
	EntityID   string `gorm:"column:entity_id;not null" json:"entityId"`
	BaseModel

	// Relationships
	Entity Entity `gorm:"foreignKey:EntityID;references:EntityID" json:"entity"`
}

// TableName sets the table name for GORM
func (Provider) TableName() string {
	return "providers"
}

// Consumer represents the consumers table
type Consumer struct {
	ConsumerID string `gorm:"primarykey;column:consumer_id" json:"consumerId"`
	EntityID   string `gorm:"column:entity_id;not null" json:"entityId"`
	BaseModel

	// Relationships
	Entity Entity `gorm:"foreignKey:EntityID;references:EntityID" json:"entity"`
}

// TableName sets the table name for GORM
func (Consumer) TableName() string {
	return "consumers"
}
