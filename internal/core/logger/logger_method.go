package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/aryanwalia/synapse/internal/core/errors"
)

// Init initializes the global logger.
func Init(level slog.Level, isJSON bool) {
	var handler slog.Handler
	if isJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	GlobalLogger = &SynapseLogger{
		InternalLogger: slog.New(handler),
	}
}

// Info logs an informational message.
func Info(ctx context.Context, msg string, attrs ...any) {
	GlobalLogger.InternalLogger.InfoContext(ctx, msg, attrs...)
}

// Error logs an error message, automatically extracting stack traces from AppError.
func Error(ctx context.Context, msg string, err error, attrs ...any) {
	allAttrs := append(attrs, slog.Any("error", err))

	if appErr, ok := err.(*errors.AppError); ok {
		allAttrs = append(allAttrs, slog.String("stack_trace", appErr.FormatStack()))
		allAttrs = append(allAttrs, slog.String("error_code", string(appErr.Code)))
	}

	GlobalLogger.InternalLogger.ErrorContext(ctx, msg, allAttrs...)
}

// Warn logs a warning message.
func Warn(ctx context.Context, msg string, attrs ...any) {
	GlobalLogger.InternalLogger.WarnContext(ctx, msg, attrs...)
}

// Debug logs a debug message.
func Debug(ctx context.Context, msg string, attrs ...any) {
	GlobalLogger.InternalLogger.DebugContext(ctx, msg, attrs...)
}
