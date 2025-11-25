package models

// Consent-engine error messages
const (
	ErrConsentNotFound     = "consent record not found"
	ErrConsentCreateFailed = "failed to create consent record"
	ErrConsentUpdateFailed = "failed to update consent record"
	ErrConsentRevokeFailed = "failed to revoke consent record"
	ErrConsentGetFailed    = "failed to get consent records"
	ErrConsentExpiryFailed = "failed to check consent expiry"
	ErrPortalRequestFailed = "failed to process consent portal request"
)

// Error codes
const (
	ErrorCodeConsentNotFound = "CONSENT_NOT_FOUND"
	ErrorCodeInternalError   = "INTERNAL_ERROR"
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeUnauthorized    = "UNAUTHORIZED"
	ErrorCodeForbidden       = "FORBIDDEN"
)

// ErrorResponseWithCode represents an error response with a code
type ErrorResponseWithCode struct {
	Code  string `json:"code,omitempty"`
	Error string `json:"error"`
}
