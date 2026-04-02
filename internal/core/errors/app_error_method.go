package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError with a stack trace.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Stack:   captureStack(3), // Skip 3 frames: runtime, captureStack, New
	}
}

// Wrap wraps an existing error into an AppError with a stack trace.
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Stack:   captureStack(3),
	}
}

// captureStack captures the current stack trace.
func captureStack(skip int) []StackFrame {
	stack := make([]StackFrame, 0)
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return stack
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		stack = append(stack, StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}
	return stack
}

// FormatStack returns a human-readable string of the stack trace.
func (e *AppError) FormatStack() string {
	var sb strings.Builder
	for _, frame := range e.Stack {
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
	}
	return sb.String()
}
