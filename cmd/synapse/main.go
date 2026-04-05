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
	"github.com/aryanwalia/synapse/internal/core/errors"
	"github.com/aryanwalia/synapse/internal/core/logger"
	"github.com/aryanwalia/synapse/internal/core/normalization"
	sqliteStore "github.com/aryanwalia/synapse/internal/data/pocketbase"
	_ "github.com/aryanwalia/synapse/internal/data/pocketbase/migrations"
	"github.com/aryanwalia/synapse/internal/ingest/pool"
	"github.com/aryanwalia/synapse/internal/ingest/sweeper"
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
	// Background services share a cancellable context
	serviceCtx, cancelServices := context.WithCancel(ctx)

	normalizationEngine := buildNormalizationEngine()

	// Normalizer dispatched by each worker
	normalizerFunc := buildNormalizerFunc(normalizationEngine, store)

	// WorkerPool: N partitions + unsorted lane, polling DB every 500ms
	workerPool := pool.NewWorkerPool(
		pool.PoolConfig{N: 8},
		store.Webhooks,
		normalizerFunc,
	)
	workerPool.Start(serviceCtx)

	// Sweeper: resets PROCESSING jobs stuck > 10 minutes
	jobSweeper := sweeper.NewSweeper(store.Webhooks, 10*time.Minute)
	go jobSweeper.Run(serviceCtx)

	// HTTP ingest pipeline (outbox write only)
	ingestPipeline := ingest.NewIngestPipeline(normalizationEngine, store.Webhooks)

	router := buildRouter(ingestPipeline)

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

	return func() {
		logger.Info(ctx, "Shutting down Synapse services...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		// Stop HTTP — no new requests
		server.Shutdown(shutdownCtx)
		// Cancel serviceCtx — pool pollers and sweeper exit gracefully
		cancelServices()
		workerPool.Stop()
		logger.Info(ctx, "Synapse services stopped cleanly")
	}
}

func buildNormalizationEngine() *normalization.NormalizationEngine {
	return normalization.NewNormalizationEngine(
		[]domain.WMSNormalizer{uniware.NewUniwareNormalizer()},
		[]domain.DISNormalizer{loginext.NewLoginextNormalizer()},
	)
}

func buildNormalizerFunc(engine *normalization.NormalizationEngine, store *sqliteStore.Storage) func(context.Context, *domain.RawWebhook) error {
	return func(ctx context.Context, wh *domain.RawWebhook) error {
		if wh.WebhookType == string(ingest.WebhookTypeDISStatusUpdate) {
			statusUpdate, err := engine.NormalizeDISStatusUpdate(ctx, wh.Source, wh.Payload)
			if err != nil {
				return err
			}
			return store.Orders.UpdateStatus(ctx, statusUpdate.ProviderAWB, string(statusUpdate.NewCanonicalStatus), wh.CorrelationID)
		}
		if wh.WebhookType == string(ingest.WebhookTypeWMSOrderCreation) {
			normalizedOrder, err := engine.NormalizeWMSOrder(ctx, wh.Source, wh.Payload)
			if err != nil {
				return err
			}
			c := normalization.MapNormalizedOrderToComposite(normalizedOrder)
			return store.Orders.CreateCompositeOrder(ctx, c.Order, c.Financials, c.Metrics, c.Recipient, c.Items)
		}
		return errors.New(errors.CodeValidation, "unknown webhook type: "+wh.WebhookType)
	}
}

func buildRouter(pipeline *ingest.IngestPipeline) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /api/v1/webhook/dis/{provider}", ingest.HandleDISWebhook(pipeline))
	router.HandleFunc("POST /api/v1/webhook/wms/{provider}", ingest.HandleWMSWebhook(pipeline))
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"synapse"}`))
	})
	return router
}
