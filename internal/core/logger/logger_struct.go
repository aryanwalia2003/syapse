package logger

import (
	"log/slog"
)

// SynapseLogger wraps the slog.Logger to provide a standard interface.
// It follows the xyz_struct.go naming convention.
type SynapseLogger struct {
	InternalLogger *slog.Logger
}

var GlobalLogger *SynapseLogger
