package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aryanwalia/synapse/internal/adapter/ingest"
	"github.com/aryanwalia/synapse/internal/core/api"
	"github.com/aryanwalia/synapse/internal/core/config"
	"github.com/aryanwalia/synapse/internal/core/logger"
	_ "github.com/aryanwalia/synapse/internal/data/pocketbase/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Configuration
	applicationConfig, err := config.Load()
	if err != nil {
		fmt.Printf("FATAL: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	isProduction := applicationConfig.Server.Env == "production"
	logger.Init(slog.LevelDebug, isProduction)

	// 3. Initialize PocketBase
	pb := pocketbase.New()

	// Register automigrations
	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{
		Automigrate: true,
		Dir:         "internal/data/pocketbase/migrations",
	})

	// Bind Synapse Server to PB Lifecycle
	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Run Synapse HTTP server in a goroutine
		stopServer := startSynapseServer(ctx, applicationConfig)

		// Ensure cleanup on PB exit
		e.App.OnTerminate().BindFunc(func(te *core.TerminateEvent) error {
			stopServer()
			return te.Next()
		})

		return e.Next()
	})

	if err := pb.Start(); err != nil {
		logger.Error(ctx, "PocketBase failed to start", err)
		os.Exit(1)
	}
}

// startSynapseServer starts the HTTP server and returns a shutdown function
func startSynapseServer(ctx context.Context, cfg *config.SynapseConfig) func() {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/v1/webhook/ingest", handleWebhookIngest)
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"synapse"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info(ctx, "Synapse HTTP Server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "Synapse HTTP server error", err)
		}
	}()

	// Graceful shutdown logic for the Synapse server
	return func() {
		logger.Info(ctx, "Shutting down Synapse HTTP Server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error(ctx, "Synapse HTTP server forced shutdown", err)
		}
		logger.Info(ctx, "Synapse HTTP Server stopped")
	}
}

func handleWebhookIngest(responseWriter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	rawPayload, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Error(ctx, "Failed to read webhook request body", err)
		http.Error(responseWriter, "Bad Request", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	vendorSource := request.Header.Get("X-Vendor-Source")
	if vendorSource == "" {
		vendorSource = "UNKNOWN"
	}

	webhookRequest := ingest.WebhookRequest{
		Source:     vendorSource,
		Payload:    rawPayload,
		Header:     request.Header,
		ReceivedAt: time.Now().UTC(),
	}

	ingestResult := ingest.ProcessIngest(ctx, webhookRequest)

	logger.Info(ctx, "Webhook ingest request completed",
		"source", webhookRequest.Source,
		"correlation_id", ingestResult.CorrelationID)

	apiResponse := api.NewSuccessResponse(ingestResult, ingestResult.CorrelationID)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusAccepted)
	json.NewEncoder(responseWriter).Encode(apiResponse)
}
