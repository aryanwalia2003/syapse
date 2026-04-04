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
	"github.com/pocketbase/pocketbase/tools/security"
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
		} else if wh.WebhookType == string(ingest.WebhookTypeWMSOrderCreation) {
			normalizedOrder, err := engine.NormalizeWMSOrder(ctx, wh.Source, wh.Payload)
			if err != nil {
				return err
			}

			orderID := "ORD-" + security.RandomString(10)
			order := &domain.Order{
				ID:              orderID,
				BrandID:         "DEFAULT_BRAND",
				WarehouseID:     "DEFAULT_WAREHOUSE",
				WMSOrderID:      normalizedOrder.ReferenceCode,
				CanonicalStatus: string(normalizedOrder.CanonicalStatus),
			}

			var codAmount int64
			if normalizedOrder.Financials.PaymentMode == domain.PaymentModeCOD {
				codAmount = normalizedOrder.Financials.TotalAmountPaise
			}

			financials := &domain.OrderFinancials{
				OrderID:        orderID,
				PaymentMode:    string(normalizedOrder.Financials.PaymentMode),
				CODAmountPaise: codAmount,
				Currency:       normalizedOrder.Financials.Currency,
			}

			metrics := &domain.OrderMetrics{
				OrderID:     orderID,
				WeightGrams: normalizedOrder.PackageWeightGrams,
				LengthCm:    normalizedOrder.Dimensions.LengthCm,
				WidthCm:     normalizedOrder.Dimensions.WidthCm,
				HeightCm:    normalizedOrder.Dimensions.HeightCm,
			}

			recipient := &domain.OrderRecipient{
				OrderID:     orderID,
				Name:        normalizedOrder.DeliveryAddress.ContactName,
				Phone:       normalizedOrder.DeliveryAddress.Phone,
				City:        normalizedOrder.DeliveryAddress.City,
				State:       normalizedOrder.DeliveryAddress.State,
				Pincode:     normalizedOrder.DeliveryAddress.PinCode,
				FullAddress: normalizedOrder.DeliveryAddress.AddressLine1 + " " + normalizedOrder.DeliveryAddress.AddressLine2,
			}

			var items []domain.OrderItem
			for _, item := range normalizedOrder.Items {
				items = append(items, domain.OrderItem{
					ID:         "ITEM-" + security.RandomString(10),
					OrderID:    orderID,
					SKU:        item.SKU,
					Name:       item.Name,
					Quantity:   item.Quantity,
					PricePaise: item.PricePaise,
				})
			}

			return store.Orders.CreateCompositeOrder(ctx, order, financials, metrics, recipient, items)
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
