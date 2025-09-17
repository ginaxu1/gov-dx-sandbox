package models

// ConsentWorkflowRequest represents a request to initiate a consent workflow
type ConsentWorkflowRequest struct {
	AppID      string      `json:"app_id"`
	DataFields []DataField `json:"data_fields"`
	Purpose    string      `json:"purpose"`
	SessionID  string      `json:"session_id"`
}

// DataField represents a data field request in the consent workflow
type DataField struct {
	OwnerType string   `json:"owner_type"`
	OwnerID   string   `json:"owner_id"`
	Fields    []string `json:"fields"`
}
