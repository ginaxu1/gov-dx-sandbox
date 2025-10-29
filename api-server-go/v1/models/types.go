package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
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

	// Try to unmarshal as array of objects first (new format)
	var records []SelectedFieldRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		// If that fails, try to unmarshal as array of strings (old format) and convert
		var stringArray []string
		if err2 := json.Unmarshal(bytes, &stringArray); err2 != nil {
			// If that also fails, try to unmarshal as a single object and wrap it
			var singleRecord SelectedFieldRecord
			if err3 := json.Unmarshal(bytes, &singleRecord); err3 != nil {
				return fmt.Errorf("cannot unmarshal JSON into SelectedFieldRecords: tried array of objects (%v), array of strings (%v), and single object (%v)", err, err2, err3)
			}
			records = []SelectedFieldRecord{singleRecord}
		} else {
			// Convert array of strings to array of objects (default schemaId to empty)
			records = make([]SelectedFieldRecord, len(stringArray))
			for i, fieldPath := range stringArray {
				records[i] = SelectedFieldRecord{
					FieldName: fieldPath,
					SchemaID:  "", // Old format doesn't have schemaId
				}
			}
		}
	}

	*sfr = SelectedFieldRecords(records)
	return nil
}

// Value implements the driver.Valuer interface for SelectedFieldRecords
func (sfr *SelectedFieldRecords) Value() (driver.Value, error) {
	return json.Marshal(*sfr)
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
