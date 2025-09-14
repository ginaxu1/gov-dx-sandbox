package models

import "github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types"

// ConsentWorkflowRequest represents a request to initiate a consent workflow
type ConsentWorkflowRequest struct {
	AppID       string            `json:"app_id"`
	DataFields  []types.DataField `json:"data_fields"`
	Purpose     string            `json:"purpose"`
	SessionID   string            `json:"session_id"`
	RedirectURL string            `json:"redirect_url"`
}
