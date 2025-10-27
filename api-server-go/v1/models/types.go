package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

//// JSONB represents a JSONB field for PostgreSQL
//type JSONB map[string]interface{}
//
//// Value returns the JSON encoding of JSONB
//func (j JSONB) Value() (driver.Value, error) {
//	if j == nil {
//		return nil, nil
//	}
//	return json.Marshal(j)
//}
//
//// Scan decodes JSON into JSONB
//func (j *JSONB) Scan(value interface{}) error {
//	if value == nil {
//		*j = nil
//		return nil
//	}
//	switch v := value.(type) {
//	case []byte:
//		return json.Unmarshal(v, j)
//	case string:
//		return json.Unmarshal([]byte(v), j)
//	default:
//		return fmt.Errorf("cannot scan %T into JSONB", value)
//	}
//}
//
//// JSONBArray represents a JSONB[] field for PostgreSQL
//type JSONBArray []map[string]interface{}
//
//// GormDataType returns the data type for GORM
//func (JSONBArray) GormDataType() string {
//	return "jsonb[]"
//}
//
//// Value returns the JSON encoding of JSONBArray
//func (j JSONBArray) Value() (driver.Value, error) {
//	if j == nil {
//		return nil, nil
//	}
//	return json.Marshal([]map[string]interface{}(j))
//}
//
//// Scan decodes PostgreSQL array into JSONBArray
//func (j *JSONBArray) Scan(value interface{}) error {
//	if value == nil {
//		*j = nil
//		return nil
//	}
//	switch v := value.(type) {
//	case []byte:
//		return json.Unmarshal(v, j)
//	case string:
//		return json.Unmarshal([]byte(v), j)
//	default:
//		return fmt.Errorf("cannot scan %T into JSONBArray", value)
//	}
//}
//
//// StringArray represents a PostgreSQL text[] field
//type StringArray []string
//
//// GormDataType returns the data type for GORM
//func (StringArray) GormDataType() string {
//	return "text[]"
//}
//
//// Value returns the string representation of StringArray for PostgreSQL
//func (s StringArray) Value() (driver.Value, error) {
//	if s == nil {
//		return nil, nil
//	}
//	return json.Marshal([]string(s))
//}
//
//// Scan decodes PostgreSQL array into StringArray
//func (s *StringArray) Scan(value interface{}) error {
//	if value == nil {
//		*s = nil
//		return nil
//	}
//
//	switch v := value.(type) {
//	case []byte:
//		var arr []string
//		err := json.Unmarshal(v, &arr)
//		if err != nil {
//			// Try PostgreSQL array format parsing
//			return s.scanPostgreSQLArray(v)
//		}
//		*s = StringArray(arr)
//		return nil
//	case string:
//		var arr []string
//		err := json.Unmarshal([]byte(v), &arr)
//		if err != nil {
//			// Try PostgreSQL array format parsing
//			return s.scanPostgreSQLArray([]byte(v))
//		}
//		*s = StringArray(arr)
//		return nil
//	default:
//		return fmt.Errorf("cannot scan %T into StringArray", value)
//	}
//}
//
//func (s *StringArray) scanPostgreSQLArray(data []byte) error {
//	str := string(data)
//	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
//		return fmt.Errorf("invalid array format: %s", str)
//	}
//
//	str = str[1 : len(str)-1] // Remove { and }
//	if str == "" {
//		*s = StringArray{}
//		return nil
//	}
//
//	parts := []string{}
//	current := ""
//	inQuotes := false
//
//	for i, char := range str {
//		if char == '"' && (i == 0 || str[i-1] != '\\') {
//			inQuotes = !inQuotes
//		} else if char == ',' && !inQuotes {
//			if current != "" {
//				parts = append(parts, s.removeQuotes(current))
//			}
//			current = ""
//		} else {
//			current += string(char)
//		}
//
//		// Handle last element
//		if i == len(str)-1 && current != "" {
//			parts = append(parts, s.removeQuotes(current))
//		}
//	}
//
//	*s = StringArray(parts)
//	return nil
//}

//// removeQuotes removes surrounding quotes and handles basic escape sequences
//func (s *StringArray) removeQuotes(str string) string {
//	// Note: This implementation only removes surrounding quotes.
//	// It does NOT handle escape sequences inside quoted strings (e.g., \" or \\).
//	// For full PostgreSQL array parsing, consider using a dedicated parser.
//	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
//		unquoted := str[1 : len(str)-1]
//		return unquoted
//	}
//	return str
//}

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

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into SelectedFieldRecords", value)
	}

	return json.Unmarshal(bytes, sfr)
}

// Value implements the driver.Valuer interface for SelectedFieldRecords
func (sfr SelectedFieldRecords) Value() (driver.Value, error) {
	if len(sfr) == 0 {
		return json.Marshal([]SelectedFieldRecord{})
	}
	return json.Marshal(sfr)
}

// PDP Data Types

// GrantDurationType GrantDuration represents the grant type enum
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
func (act AccessControlType) Value() (driver.Value, error) {
	return string(act), nil
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
func (s Source) Value() (driver.Value, error) {
	return string(s), nil
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
func (o Owner) Value() (driver.Value, error) {
	return string(o), nil
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
func (al AllowList) Value() (driver.Value, error) {
	if len(al) == 0 {
		return json.Marshal(map[string]AllowListEntry{})
	}
	return json.Marshal(al)
}
