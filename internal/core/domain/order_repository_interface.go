package domain

import "context"

type OrderRepository interface {
	CreateCompositeOrder(
		ctx context.Context,
		order *Order,
		financials *OrderFinancials,
		metrics *OrderMetrics,
		recipient *OrderRecipient,
		items []OrderItem,
	) error

	GetByID(ctx context.Context, id string) (*Order, error)

	UpdateStatus(ctx context.Context, orderID string, newStatus string, correlationID string) error
}
