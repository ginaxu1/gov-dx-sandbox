package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB represents a JSONB field for PostgreSQL
type JSONB map[string]interface{}

// Value returns the JSON encoding of JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan decodes JSON into JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// StringArray represents a PostgreSQL text[] field
type StringArray []string

// GormDataType returns the data type for GORM
func (StringArray) GormDataType() string {
	return "text[]"
}

// Value returns the string representation of StringArray for PostgreSQL
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal([]string(s))
}

// Scan decodes PostgreSQL array into StringArray
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		var arr []string
		err := json.Unmarshal(v, &arr)
		if err != nil {
			// Try PostgreSQL array format parsing
			return s.scanPostgreSQLArray(v)
		}
		*s = StringArray(arr)
		return nil
	case string:
		var arr []string
		err := json.Unmarshal([]byte(v), &arr)
		if err != nil {
			// Try PostgreSQL array format parsing
			return s.scanPostgreSQLArray([]byte(v))
		}
		*s = StringArray(arr)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

func (s *StringArray) scanPostgreSQLArray(data []byte) error {
	str := string(data)
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("invalid array format: %s", str)
	}

	str = str[1 : len(str)-1] // Remove { and }
	if str == "" {
		*s = StringArray{}
		return nil
	}

	// Simple parsing - assumes no commas in values
	parts := []string{}
	current := ""
	inQuotes := false

	for i, char := range str {
		if char == '"' {
			inQuotes = !inQuotes
		} else if char == ',' && !inQuotes {
			if current != "" {
				// Remove surrounding quotes if present
				if len(current) >= 2 && current[0] == '"' && current[len(current)-1] == '"' {
					current = current[1 : len(current)-1]
				}
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(char)
		}

		// Handle last element
		if i == len(str)-1 && current != "" {
			if len(current) >= 2 && current[0] == '"' && current[len(current)-1] == '"' {
				current = current[1 : len(current)-1]
			}
			parts = append(parts, current)
		}
	}

	*s = StringArray(parts)
	return nil
}
