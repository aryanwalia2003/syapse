package errors

// ErrorCode defines a standard set of error codes for the application.
type ErrorCode string

const (
	CodeInternal     ErrorCode = "INTERNAL"
	CodeValidation   ErrorCode = "VALIDATION"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeConflict     ErrorCode = "CONFLICT"
)

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// AppError is the standard error type for Project Synapse.
// It follows the xyz_struct.go naming convention.
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
	Stack   []StackFrame
}
