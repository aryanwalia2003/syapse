package api

import (
	"github.com/aryanwalia/synapse/internal/core/errors"
)

// NewSuccessResponse creates a successful API response envelope.
func NewSuccessResponse[T any](data T, correlationID string) Response[T] {
	return Response[T]{
		Success:       true,
		CorrelationID: correlationID,
		Data:          data,
	}
}

// NewErrorResponse creates an error API response envelope by mapping an internal AppError.
func NewErrorResponse(err error, correlationID string) Response[any] {
	apiErr := &ApiError{
		Code:    string(errors.CodeInternal),
		Message: "An unexpected error occurred",
	}

	if appErr, ok := err.(*errors.AppError); ok {
		apiErr.Code = string(appErr.Code)
		apiErr.Message = appErr.Message
	}

	return Response[any]{
		Success:       false,
		CorrelationID: correlationID,
		Error:         apiErr,
	}
}
