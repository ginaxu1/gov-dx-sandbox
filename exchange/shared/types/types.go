// Package types provides shared data structures across services
package types

// DataField represents a data owner and their fields
type DataField struct {
	// OwnerType indicates the type of data owner (e.g., "citizen", "organization")
	OwnerType string `json:"owner_type"`
	// OwnerID is the unique identifier for the data owner
	OwnerID string `json:"owner_id"`
	// Fields is the list of specific data fields for this owner
	Fields []string `json:"fields"`
}
