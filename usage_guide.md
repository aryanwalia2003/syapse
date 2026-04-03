# Synapse Development Guide: Building a New API

This guide explains how to use our decoupled, SQL-first architecture to add new features or APIs to Synapse Logistics Middleware.

---

## The "Synapse Way" (Architecture Summary)
We use a **4-Layer Architecture** to ensure that switching databases (e.g., SQLite to PostgreSQL) or WMS providers (e.g., EasyEcom to Unicommerce) requires zero changes to the core logic.

1.  **Domain Layer** (`internal/core/domain`): The "Source of Truth" (Interfaces & Structs).
2.  **Data Layer** (`internal/data/sql`): Raw SQL implementations.
3.  **Database Wrapper** (`internal/data/pocketbase`): Bridge between PB and our SQL layer.
4.  **Adapter Layer** (`internal/adapter`): HTTP Handlers and External Integrations.

---

## Step 1: Define the Domain Interface
Add your new method to the appropriate interface in `internal/core/domain/`.

```go
// internal/core/domain/order_repository_interface.go
type OrderRepository interface {
    // ... existing methods
    FindAll(ctx context.Context, brandID string) ([]Order, error)
}
```

---

## Step 2: Implement the SQL Logic
Go to the corresponding component folder in `internal/data/sql/`. Create a new file for the method following the `[action]_method.go` nomenclature.

```go
// internal/data/sql/order/find_all_method.go
package order

func (repo *SQLOrderRepository) FindAll(ctx context.Context, brandID string) ([]domain.Order, error) {
    query := `SELECT * FROM orders WHERE brand_id = {:brand_id} ORDER BY created_at DESC`
    
    var orders []domain.Order
    // Use .QueryRows for multiple results, .QueryRow for single, .Execute for Inserts/Updates
    err := repo.db.QueryRows(ctx, query, map[string]any{"brand_id": brandID}, &orders)
    
    return orders, err
}
```

---

## Step 3: Create the HTTP Handler
Create a new handler in `internal/adapter/`. Use **Dependency Injection** by passing the required repository interface.

```go
func HandleGetOrders(orderRepo domain.OrderRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        brandID := r.URL.Query().Get("brand_id")
        
        orders, err := orderRepo.FindAll(r.Context(), brandID)
        if err != nil {
            // Use internal/core/api to send errors
            return
        }
        
        // Return Success Response
    }
}
```

---

## Step 4: Wire it up in `main.go`
In `cmd/synapse/main.go`, register your new route and inject the repository from the `Storage` container.

```go
// Inside startSynapseServer function
router.HandleFunc("GET /api/v1/orders", HandleGetOrders(store.Orders))
```

---

## Key Utility: `coreDB.DB`
Never import PocketBase directly into your SQL files. Always use the `db` interface:
- `Execute(ctx, sql, params)`: For `INSERT`, `UPDATE`, `DELETE`.
- `QueryRow(ctx, sql, params, &dest)`: For single `SELECT`.
- `QueryRows(ctx, sql, params, &destSlice)`: For multiple `SELECT`.
- `RunInTransaction(ctx, func(tx) error)`: To group multiple queries safely.

## Named Parameters
Always use `{:parameter_name}` in your SQL strings. Our wrapper automatically maps these to the underlying database engine, preventing SQL Injection.

```go
query := "UPDATE orders SET status = {:s} WHERE id = {:id}"
params := map[string]any{"s": "SHIPPED", "id": "123"}
```
