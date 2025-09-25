package errors

import (
	"database/sql"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeDatabase     ErrorType = "database"
	ErrorTypeTimeout      ErrorType = "timeout"
	ErrorTypeNetwork      ErrorType = "network"
)

// APIError represents a structured API error
type APIError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	HTTPStatus  int       `json:"-"`
	InternalErr error     `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Message, e.Details, e.Code)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.InternalErr
}

// NewAPIError creates a new API error
func NewAPIError(errorType ErrorType, code, message string, httpStatus int) *APIError {
	return &APIError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(errorType ErrorType, code, message, details string, httpStatus int) *APIError {
	return &APIError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
	}
}

// NewAPIErrorWithCause creates a new API error with an underlying cause
func NewAPIErrorWithCause(errorType ErrorType, code, message string, httpStatus int, cause error) *APIError {
	return &APIError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		HTTPStatus:  httpStatus,
		InternalErr: cause,
	}
}

// Predefined error constructors

// ValidationError creates a validation error
func ValidationError(code, message string) *APIError {
	return NewAPIError(ErrorTypeValidation, code, message, http.StatusBadRequest)
}

// ValidationErrorWithDetails creates a validation error with details
func ValidationErrorWithDetails(code, message, details string) *APIError {
	return NewAPIErrorWithDetails(ErrorTypeValidation, code, message, details, http.StatusBadRequest)
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *APIError {
	return NewAPIError(ErrorTypeNotFound, "RESOURCE_NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// ConflictError creates a conflict error
func ConflictError(message string) *APIError {
	return NewAPIError(ErrorTypeConflict, "RESOURCE_CONFLICT", message, http.StatusConflict)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *APIError {
	return NewAPIError(ErrorTypeUnauthorized, "UNAUTHORIZED", message, http.StatusUnauthorized)
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *APIError {
	return NewAPIError(ErrorTypeForbidden, "FORBIDDEN", message, http.StatusForbidden)
}

// InternalError creates an internal server error
func InternalError(message string) *APIError {
	return NewAPIError(ErrorTypeInternal, "INTERNAL_ERROR", message, http.StatusInternalServerError)
}

// InternalErrorWithCause creates an internal server error with cause
func InternalErrorWithCause(message string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrorTypeInternal, "INTERNAL_ERROR", message, http.StatusInternalServerError, cause)
}

// DatabaseError creates a database error
func DatabaseError(operation string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrorTypeDatabase, "DATABASE_ERROR",
		fmt.Sprintf("Database operation failed: %s", operation),
		http.StatusInternalServerError, cause)
}

// TimeoutError creates a timeout error
func TimeoutError(operation string) *APIError {
	return NewAPIError(ErrorTypeTimeout, "TIMEOUT",
		fmt.Sprintf("Operation timed out: %s", operation),
		http.StatusRequestTimeout)
}

// NetworkError creates a network error
func NetworkError(operation string, cause error) *APIError {
	return NewAPIErrorWithCause(ErrorTypeNetwork, "NETWORK_ERROR",
		fmt.Sprintf("Network operation failed: %s", operation),
		http.StatusServiceUnavailable, cause)
}

// Error handling utilities

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError extracts APIError from an error
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

// WrapError wraps a generic error into an APIError
func WrapError(err error, errorType ErrorType, code, message string, httpStatus int) *APIError {
	return NewAPIErrorWithCause(errorType, code, message, httpStatus, err)
}

// HandleDatabaseError handles database-specific errors
func HandleDatabaseError(err error, operation string) *APIError {
	if err == nil {
		return nil
	}

	switch err {
	case sql.ErrNoRows:
		return NotFoundError("resource")
	case sql.ErrConnDone:
		return DatabaseError(operation, err)
	case sql.ErrTxDone:
		return DatabaseError(operation, err)
	default:
		// Check for PostgreSQL specific errors
		if pgErr := extractPostgreSQLError(err); pgErr != nil {
			return handlePostgreSQLError(pgErr, operation)
		}
		return DatabaseError(operation, err)
	}
}

// extractPostgreSQLError extracts PostgreSQL error information
func extractPostgreSQLError(err error) map[string]string {
	// This is a simplified version - in a real implementation,
	// you'd use a PostgreSQL driver that provides error codes
	return nil
}

// handlePostgreSQLError handles PostgreSQL specific errors
func handlePostgreSQLError(pgErr map[string]string, operation string) *APIError {
	// This is a simplified version - in a real implementation,
	// you'd check specific PostgreSQL error codes
	return DatabaseError(operation, fmt.Errorf("postgresql error"))
}

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(apiErr *APIError) *ErrorResponse {
	return &ErrorResponse{
		Error:     apiErr,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

// getCurrentTimestamp returns current timestamp (simplified)
func getCurrentTimestamp() int64 {
	// In a real implementation, you'd use time.Now().Unix()
	return 0
}
