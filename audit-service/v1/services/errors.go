package services

import "errors"

// ErrValidation represents a validation error in the domain layer
// This is a domain-specific error that abstracts away database implementation details
var ErrValidation = errors.New("validation error")

// ErrInvalidInput represents an input validation error
var ErrInvalidInput = errors.New("invalid input")

// IsValidationError checks if an error is a validation error or invalid input
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation) || errors.Is(err, ErrInvalidInput)
}
