package models

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
