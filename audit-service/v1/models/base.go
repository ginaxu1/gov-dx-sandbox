package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
// Note: UpdatedAt is intentionally omitted as audit logs are immutable (created only, never updated)
type BaseModel struct {
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// BeforeCreate GORM hook for BaseModel
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.CreatedAt = time.Now().UTC()
	return nil
}
