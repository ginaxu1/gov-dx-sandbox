package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PolicyMetadata represents the policy_metadata table
type PolicyMetadata struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SchemaID          string            `gorm:"type:varchar(255);not null;constraint:fk_policy_metadata_provider_schemas,references provider_schemas(schema_id)" json:"schema_id"`
	FieldName         string            `gorm:"type:text;not null" json:"field_name"`
	DisplayName       *string           `gorm:"type:text" json:"display_name,omitempty"`
	Description       *string           `gorm:"type:text" json:"description,omitempty"`
	Source            Source            `gorm:"type:source_enum;not null;default:'fallback'" json:"source"`
	IsOwner           bool              `gorm:"default:false;not null" json:"is_owner"`
	AccessControlType AccessControlType `gorm:"type:access_control_type_enum;not null;default:'restricted'" json:"access_control_type"`
	AllowList         AllowList         `gorm:"type:jsonb;not null;default:'{}'" json:"allow_list"`
	Owner             *Owner            `gorm:"type:owner_enum;default:'citizen'" json:"owner"`
	CreatedAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	UpdatedAt         time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (PolicyMetadata) TableName() string {
	return "policy_metadata"
}

// BeforeCreate sets the default values before creating a record
func (pm *PolicyMetadata) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	if pm.Source == "" {
		pm.Source = SourceFallback
	}
	if pm.AccessControlType == "" {
		pm.AccessControlType = AccessControlTypeRestricted
	}
	if pm.Owner == nil {
		owner := OwnerCitizen
		pm.Owner = &owner
	}
	if pm.AllowList == nil {
		pm.AllowList = make(AllowList)
	}
	now := time.Now()
	pm.CreatedAt = now
	pm.UpdatedAt = now
	return nil
}

// BeforeUpdate sets the updated_at timestamp before updating a record
func (pm *PolicyMetadata) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	pm.UpdatedAt = now
	return nil
}
