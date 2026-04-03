package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
	sqliteStore "github.com/aryanwalia/synapse/internal/data/pocketbase"
	_ "github.com/aryanwalia/synapse/internal/data/pocketbase/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/security"
)

func main() {
	logger.Init(slog.LevelDebug, false)
	ctx := context.Background()

	// 1. Initialize PocketBase in test mode (temporary directory)
	testDataDir := "./test_pb_data"
	os.RemoveAll(testDataDir)
	defer os.RemoveAll(testDataDir)

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: testDataDir,
	})

	// 2. Bootstrap (this runs migrations automatically)
	if err := app.Bootstrap(); err != nil {
		fmt.Printf("Bootstrap Failed: %v\n", err)
		os.Exit(1)
	}

	logger.Info(ctx, "🚀 Starting Synchronous Persistence Test...")

	// 3. Initialize Storage
	dbConn := sqliteStore.NewPBDatabase(app)
	store := sqliteStore.NewStorage(app)

	// 4. Setup
	brandID := security.RandomString(15)
	warehouseID := security.RandomString(15)

	_ = dbConn.Execute(ctx, "INSERT INTO brands (id, name, is_active) VALUES ({:id}, 'Test', 1)", map[string]any{"id": brandID})
	_ = dbConn.Execute(ctx, "INSERT INTO warehouses (id, brand_id, facility_code, name, is_active) VALUES ({:id}, {:bid}, 'F1', 'W1', 1)",
		map[string]any{"id": warehouseID, "bid": brandID})

	// 5. Create Composite Order
	order := &domain.Order{
		ID:              security.RandomString(15),
		BrandID:         brandID,
		WarehouseID:     warehouseID,
		WMSOrderID:      "TEST-123",
		CanonicalStatus: "ACCEPTED",
	}
	items := []domain.OrderItem{{SKU: "TEST-SKU", Name: "Test Item", Quantity: 1, PricePaise: 100}}

	err := store.Orders.CreateCompositeOrder(ctx, order, nil, nil, nil, items)
	if err != nil {
		logger.Error(ctx, "❌ Save Failed", err)
		os.Exit(1)
	}

	// 6. Verify
	fetched, err := store.Orders.GetByID(ctx, order.ID)
	if err != nil || fetched == nil || fetched.WMSOrderID != "TEST-123" {
		logger.Error(ctx, "❌ Verify Failed", err)
		os.Exit(1)
	}

	logger.Info(ctx, "✨ INTEGRATION TEST PASSED (Synchronous) ✨")
}
