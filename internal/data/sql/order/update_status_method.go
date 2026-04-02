package order

import (
	"context"

	"github.com/aryanwalia/synapse/internal/core/db"
	"github.com/pocketbase/pocketbase/tools/security"
)

func (repo *SQLOrderRepository) UpdateStatus(ctx context.Context, orderID string, newStatus string, correlationID string) error {
	return repo.db.RunInTransaction(ctx, func(tx db.Transaction) error {
		statusSQL := `UPDATE orders SET canonical_status = {:status} WHERE id = {:id}`
		err := tx.Execute(ctx, statusSQL, map[string]any{
			"status": newStatus,
			"id":     orderID,
		})
		if err != nil {
			return err
		}

		logSQL := `
			INSERT INTO order_state_logs (id, order_id, correlation_id, new_status, triggered_by)
			VALUES ({:id}, {:order_id}, {:correlation_id}, {:new_status}, 'SYSTEM')
		`
		return tx.Execute(ctx, logSQL, map[string]any{
			"id":             security.RandomString(15),
			"order_id":       orderID,
			"correlation_id": correlationID,
			"new_status":     newStatus,
		})
	})
}
