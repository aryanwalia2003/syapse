package pocketbase

import (
	"github.com/aryanwalia/synapse/internal/core/domain" // Optional: if wrapping DB here
	"github.com/aryanwalia/synapse/internal/data/sql/order"
	"github.com/aryanwalia/synapse/internal/data/sql/webhook"
	"github.com/pocketbase/pocketbase/core"
)

// Storage acts as a container for all our repositories.
type Storage struct {
	Orders   domain.OrderRepository
	Webhooks domain.WebhookRepository
}

// NewStorage initializes all repositories using the provided PocketBase app.
func NewStorage(app core.App) *Storage {
	dbConn := NewPBDatabase(app)

	return &Storage{
		Orders:   order.NewSQLOrderRepository(dbConn),
		Webhooks: webhook.NewSQLWebhookRepository(dbConn),
	}
}
