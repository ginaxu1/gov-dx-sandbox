package errors

// PDP-related
const (
	CodePDPNotAllowed  = "PDP_NOT_ALLOWED"
	CodePDPUnavailable = "PDP_UNAVAILABLE"
	CodePDPError       = "PDP_ERROR"
	CodePDPNoResponse  = "PDP_NO_RESPONSE"
)

// CE-related
const (
	CodeCEUnavailable    = "CE_UNAVAILABLE"
	CodeCEError          = "CE_ERROR"
	CodeCENoResponse     = "CE_NO_RESPONSE"
	CodeCEConsentDenied  = "CE_CONSENT_DENIED"
	CodeCEConsentExpired = "CE_CONSENT_EXPIRED"
	CodeCENotApproved    = "CE_NOT_APPROVED"
)

// OE-related
const (
	CodeMissingEntityIdentifier = "MISSING_IDENTIFIER"
)

// Auth-related
const (
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
)

// Generic
const (
	CodeInternalError = "INTERNAL_ERROR"
	CodeBadRequest    = "BAD_REQUEST"
)
