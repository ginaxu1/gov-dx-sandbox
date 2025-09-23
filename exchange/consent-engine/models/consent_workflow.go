package models

// ConsentWorkflowRequest represents a request to initiate a consent workflow
type ConsentWorkflowRequest struct {
	AppID      string      `json:"app_id"`
	DataFields []DataField `json:"data_fields"`
	SessionID  string      `json:"session_id"`
}

// DataField represents a data field request in the consent workflow
type DataField struct {
	OwnerType  string   `json:"owner_type"`
	OwnerID    string   `json:"owner_id"`
	OwnerEmail string   `json:"owner_email"`
	Fields     []string `json:"fields"`
}
