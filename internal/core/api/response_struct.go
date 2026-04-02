package api

// ApiError represents the external-facing error structure.
// Following the xyz_struct.go naming convention.
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response is the universal API envelope for all synapse responses.
type Response[T any] struct {
	Success       bool      `json:"success"`
	CorrelationID string    `json:"correlation_id"`
	Data          T         `json:"data,omitempty"`
	Error         *ApiError `json:"error,omitempty"`
}
