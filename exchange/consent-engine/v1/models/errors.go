package models

// ErrorResponseWithCode represents an error response with a code
type ErrorResponseWithCode struct {
	Code  string `json:"code,omitempty"`
	Error string `json:"error"`
}
