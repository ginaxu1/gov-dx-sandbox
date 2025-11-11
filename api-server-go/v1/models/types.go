package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SelectedFieldRecord represents a record in the selected_fields array
type SelectedFieldRecord struct {
	FieldName string `json:"fieldName"`
	SchemaID  string `json:"schemaId"`
}

// SelectedFieldRecords represents an array of SelectedFieldRecord with custom scanning
type SelectedFieldRecords []SelectedFieldRecord

// Scan implements the sql.Scanner interface for SelectedFieldRecords
func (sfr *SelectedFieldRecords) Scan(value interface{}) error {
	if value == nil {
		*sfr = SelectedFieldRecords{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into SelectedFieldRecords", value)
	}

	return json.Unmarshal(bytes, sfr)
}

// Value implements the driver.Valuer interface for SelectedFieldRecords
func (sfr *SelectedFieldRecords) Value() (driver.Value, error) {
	return json.Marshal(*sfr)
}

// GormDataType gorm common data type
func (SelectedFieldRecords) GormDataType() string {
	return "jsonb"
}

// GormValue implements the GormValuerInterface
func (sfr SelectedFieldRecords) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, err := json.Marshal(sfr)
	if err != nil {
		// Panic on marshaling error to prevent silent data loss
		// JSON marshaling of SelectedFieldRecords should never fail under normal circumstances
		panic(fmt.Sprintf("Failed to marshal SelectedFieldRecords to JSON: %v", err))
	}

	// Detect database type for SQLite compatibility
	dialector := db.Dialector.Name()
	var sqlExpr string
	if dialector == "sqlite" {
		// SQLite uses TEXT for JSON, no cast needed
		sqlExpr = "?"
	} else {
		// PostgreSQL uses jsonb with cast
		sqlExpr = "?::jsonb"
	}

	return clause.Expr{
		SQL:  sqlExpr,
		Vars: []interface{}{string(data)},
	}
}

// PDP Data Types

// GrantDurationType represents the grant duration type enum
type GrantDurationType string

const (
	GrantDurationTypeOneMonth GrantDurationType = "30d"
	GrantDurationTypeOneYear  GrantDurationType = "365d"
)

// AccessControlType represents the access control type enum
type AccessControlType string

const (
	AccessControlTypePublic     AccessControlType = "public"
	AccessControlTypeRestricted AccessControlType = "restricted"
)

// Scan implements the sql.Scanner interface for AccessControlType
func (act *AccessControlType) Scan(value interface{}) error {
	if value == nil {
		*act = AccessControlTypeRestricted
		return nil
	}
	if str, ok := value.(string); ok {
		*act = AccessControlType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into AccessControlType", value)
}

// Value implements the driver.Valuer interface for AccessControlType
func (act *AccessControlType) Value() (driver.Value, error) {
	return string(*act), nil
}

// Source represents the source enum
type Source string

const (
	SourcePrimary  Source = "primary"
	SourceFallback Source = "fallback"
)

// Scan implements the sql.Scanner interface for Source
func (s *Source) Scan(value interface{}) error {
	if value == nil {
		*s = SourceFallback
		return nil
	}
	if str, ok := value.(string); ok {
		*s = Source(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Source", value)
}

// Value implements the driver.Valuer interface for Source
func (s *Source) Value() (driver.Value, error) {
	return string(*s), nil
}

// Owner represents the owner enum
type Owner string

const (
	OwnerCitizen Owner = "citizen"
)

// Scan implements the sql.Scanner interface for Owner
func (o *Owner) Scan(value interface{}) error {
	if value == nil {
		*o = OwnerCitizen
		return nil
	}
	if str, ok := value.(string); ok {
		*o = Owner(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Owner", value)
}

// Value implements the driver.Valuer interface for Owner
func (o *Owner) Value() (driver.Value, error) {
	return string(*o), nil
}

// AllowListEntry represents an entry in the allow list
type AllowListEntry struct {
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AllowList represents the JSONB allow list as a HashMap with custom scanning
// Key: application_id, Value: AllowListEntry
type AllowList map[string]AllowListEntry

// Scan implements the sql.Scanner interface for AllowList
func (al *AllowList) Scan(value interface{}) error {
	if value == nil {
		*al = make(AllowList)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into AllowList", value)
	}

	return json.Unmarshal(bytes, al)
}

// Value implements the driver.Valuer interface for AllowList
func (al *AllowList) Value() (driver.Value, error) {
	return json.Marshal(*al)
}
