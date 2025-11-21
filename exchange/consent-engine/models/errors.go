package models

// Error codes for consent-engine API responses
// These codes align with orchestration-engine error code patterns
const (
	// Consent not found
	ErrorCodeConsentNotFound = "CONSENT_NOT_FOUND"

	// Consent validation errors
	ErrorCodeInvalidAction      = "INVALID_ACTION"
	ErrorCodeInvalidStatus       = "INVALID_STATUS"
	ErrorCodeInvalidRequest      = "INVALID_REQUEST"
	ErrorCodeMissingRequiredField = "MISSING_REQUIRED_FIELD"

	// Consent state errors
	ErrorCodeConsentExpired  = "CONSENT_EXPIRED"
	ErrorCodeConsentRevoked  = "CONSENT_REVOKED"
	ErrorCodeConsentRejected = "CONSENT_REJECTED"

	// Service errors
	ErrorCodeInternalError   = "INTERNAL_ERROR"
	ErrorCodeDatabaseError   = "DATABASE_ERROR"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// ErrorResponseWithCode represents an error response with a standardized error code
type ErrorResponseWithCode struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code,omitempty"`
}

