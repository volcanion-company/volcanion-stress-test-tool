package domain

import "errors"

// Domain errors for business logic
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates the resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrAlreadyRunning indicates a test is already running
	ErrAlreadyRunning = errors.New("test is already running")

	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input")

	// ErrValidationFailed indicates validation failed
	ErrValidationFailed = errors.New("validation failed")

	// ErrOperationFailed indicates a general operation failure
	ErrOperationFailed = errors.New("operation failed")

	// ErrSLAViolation indicates SLA thresholds were violated
	ErrSLAViolation = errors.New("SLA violation")
)

// ValidationError represents a detailed validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	return e.Resource + " not found: " + e.ID
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}
