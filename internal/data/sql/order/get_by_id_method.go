package order

import (
	"context"
	"database/sql"

	"github.com/aryanwalia/synapse/internal/core/domain"
)

func (repo *SQLOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `SELECT id, brand_id, warehouse_id, wms_order_id, canonical_status FROM orders WHERE id = {:id} LIMIT 1`

	order := &domain.Order{}
	err := repo.db.QueryRow(ctx, query, map[string]any{"id": id}, order)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}
