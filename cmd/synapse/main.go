package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aryanwalia/synapse/internal/adapter/dis/loginext"
	"github.com/aryanwalia/synapse/internal/adapter/ingest"
	"github.com/aryanwalia/synapse/internal/adapter/wms/uniware"
	"github.com/aryanwalia/synapse/internal/core/config"
	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
	"github.com/aryanwalia/synapse/internal/core/normalization"
	sqliteStore "github.com/aryanwalia/synapse/internal/data/pocketbase"
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

	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Initialize the Storage Container
		store := sqliteStore.NewStorage(e.App)

		// Run Synapse HTTP server in a goroutine
		stopServer := startSynapseServer(ctx, applicationConfig, store)

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

func startSynapseServer(ctx context.Context, cfg *config.SynapseConfig, store *sqliteStore.Storage) func() {
	router := http.NewServeMux()

	normalizationEngine := normalization.NewNormalizationEngine(
		[]domain.WMSNormalizer{
			uniware.NewUniwareNormalizer(),
		},
		[]domain.DISNormalizer{
			loginext.NewLoginextNormalizer(),
		},
	)

	ingestPipeline := ingest.NewIngestPipeline(normalizationEngine, store.Webhooks)

	// DIS status updates (e.g. Loginext, Valkyrie)
	router.HandleFunc("POST /api/v1/webhook/dis/{provider}", ingest.HandleDISWebhook(ingestPipeline))
	// WMS order creation (e.g. Uniware, EasyEcom)
	router.HandleFunc("POST /api/v1/webhook/wms/{provider}", ingest.HandleWMSWebhook(ingestPipeline))

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
