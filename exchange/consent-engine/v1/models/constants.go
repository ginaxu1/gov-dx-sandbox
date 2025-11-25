package models

// ConsentStatus represents the status of a consent record
type ConsentStatus string

// ConsentStatus constants
const (
	StatusPending  ConsentStatus = "pending"
	StatusApproved ConsentStatus = "approved"
	StatusRejected ConsentStatus = "rejected"
	StatusExpired  ConsentStatus = "expired"
	StatusRevoked  ConsentStatus = "revoked"
)

// DefaultPendingTimeoutDuration is the default duration for pending consent expiry
// Pending consents will expire after this duration if not approved or rejected
// Format: ISO 8601 duration (e.g., "P1D" for 1 day, "PT24H" for 24 hours)
const DefaultPendingTimeoutDuration = "P1D" // 1 day default

// ConsentErrorMessage represents an error message
type ConsentErrorMessage string

// ConsentErrorMessage constants
const (
	ErrConsentNotFound     ConsentErrorMessage = "consent record not found"
	ErrConsentCreateFailed ConsentErrorMessage = "failed to create consent record"
	ErrConsentUpdateFailed ConsentErrorMessage = "failed to update consent record"
	ErrConsentRevokeFailed ConsentErrorMessage = "failed to revoke consent record"
	ErrConsentGetFailed    ConsentErrorMessage = "failed to get consent records"
	ErrConsentExpiryFailed ConsentErrorMessage = "failed to check consent expiry"
	ErrPortalRequestFailed ConsentErrorMessage = "failed to process consent portal request"
)

// ConsentErrorCode represents an error code
type ConsentErrorCode string

// ConsentErrorCode constants
const (
	ErrorCodeConsentNotFound ConsentErrorCode = "CONSENT_NOT_FOUND"
	ErrorCodeInternalError   ConsentErrorCode = "INTERNAL_ERROR"
	ErrorCodeBadRequest      ConsentErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized    ConsentErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden       ConsentErrorCode = "FORBIDDEN"
)

// ConsentEngineOperation represents the operation
type ConsentEngineOperation string

// ConsentStatus constants
const (
	StatusPending  ConsentStatus = "pending"
	StatusApproved ConsentStatus = "approved"
	StatusRejected ConsentStatus = "rejected"
	StatusExpired  ConsentStatus = "expired"
	StatusRevoked  ConsentStatus = "revoked"
)

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

// ConsentEngineOperation constants
const (
	OpCreateConsent         ConsentEngineOperation = "create consent"
	OpUpdateConsent         ConsentEngineOperation = "update consent"
	OpRevokeConsent         ConsentEngineOperation = "revoke consent"
	OpGetConsentStatus      ConsentEngineOperation = "get consent status"
	OpGetConsentsByOwner    ConsentEngineOperation = "get consents by data owner"
	OpGetConsentsByConsumer ConsentEngineOperation = "get consents by consumer"
	OpCheckConsentExpiry    ConsentEngineOperation = "check consent expiry"
	OpProcessPortalRequest  ConsentEngineOperation = "process consent portal"
)
