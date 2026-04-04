package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aryanwalia/synapse/internal/core/api"
	"github.com/aryanwalia/synapse/internal/core/config"
	"github.com/aryanwalia/synapse/internal/core/errors"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func main() {
	logger.Init(slog.LevelDebug, false)

	ctx := context.Background()
	logger.Info(ctx, "Starting Configuration, API, and Webhook demonstration")

	cfg, err := config.Load()
	if err != nil {
		logger.Error(ctx, "Configuration load failed", err)
	} else {
		logger.Info(ctx, "Configuration loaded successfully", "port", cfg.Server.Port, "pb_url", cfg.Database.PocketBaseURL)
	}

	successResp := api.NewSuccessResponse(map[string]string{"status": "processed"}, "debug-1234")
	printJSON("Success Response", successResp)

	appErr := errors.New(errors.CodeValidation, "invalid signature detected")
	errorResp := api.NewErrorResponse(appErr, "debug-1234")
	printJSON("Error Response", errorResp)
}

func printJSON(label string, v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Printf("\n--- %s ---\n%s\n", label, string(b))
}
