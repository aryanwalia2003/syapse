package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aryanwalia/synapse/internal/adapter/ingest"
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

	webhookReq := ingest.WebhookRequest{
		Source:  "Loginext",
		Payload: []byte(`{"event": "delivered", "order_id": "ORD-123"}`),
	}
	ingestResult := ingest.ProcessIngest(ctx, webhookReq)
	logger.Info(ctx, "Webhook ingested", "correlation_id", ingestResult.CorrelationID)

	successResp := api.NewSuccessResponse(map[string]string{"status": "processed"}, ingestResult.CorrelationID)
	printJSON("Success Response", successResp)

	appErr := errors.New(errors.CodeValidation, "invalid signature detected")
	errorResp := api.NewErrorResponse(appErr, ingestResult.CorrelationID)
	printJSON("Error Response", errorResp)
}

func printJSON(label string, v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Printf("\n--- %s ---\n%s\n", label, string(b))
}
