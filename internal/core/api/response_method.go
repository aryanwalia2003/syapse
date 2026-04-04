package api

import (
	"encoding/json"
	"net/http"

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

// WriteSuccess writes a JSON success response to the http.ResponseWriter
func WriteSuccess[T any](w http.ResponseWriter, r *http.Request, data T, correlationID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // As per PRD 202 Accepted
	json.NewEncoder(w).Encode(NewSuccessResponse(data, correlationID))
}

// WriteError writes a JSON error response to the http.ResponseWriter
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Map some errors to specific HTTP codes
	status := http.StatusInternalServerError
	if appErr, ok := err.(*errors.AppError); ok {
		if appErr.Code == errors.CodeValidation {
			status = http.StatusBadRequest
		}
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(NewErrorResponse(err, ""))
}
