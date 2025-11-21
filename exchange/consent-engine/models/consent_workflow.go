package models

// ConsentWorkflowRequest represents a request to initiate a consent workflow
type ConsentWorkflowRequest struct {
	AppID      string      `json:"app_id"`
	DataFields []DataField `json:"data_fields"`
	SessionID  string      `json:"session_id"`
}


