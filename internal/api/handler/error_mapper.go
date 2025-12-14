package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Field   string `json:"field,omitempty"`
}

// MapErrorToHTTP maps domain errors to HTTP status codes and responses
func MapErrorToHTTP(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check for specific domain error types
	var notFoundErr *domain.NotFoundError
	var validationErr *domain.ValidationError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: notFoundErr.Error(),
		})

	case errors.As(err, &validationErr):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: validationErr.Message,
			Field:   validationErr.Field,
		})

	case errors.Is(err, domain.ErrAlreadyRunning):
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "conflict",
			Message: "Test is already running",
		})

	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "conflict",
			Message: "Resource already exists",
		})

	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_input",
			Message: err.Error(),
		})

	case errors.Is(err, domain.ErrValidationFailed):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})

	case errors.Is(err, domain.ErrSLAViolation):
		c.JSON(http.StatusExpectationFailed, ErrorResponse{
			Error:   "sla_violation",
			Message: err.Error(),
		})

	default:
		// Generic internal server error
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "An unexpected error occurred",
		})
	}
}
