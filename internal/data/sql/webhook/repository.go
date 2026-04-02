package webhook

import (
	coreDB "github.com/aryanwalia/synapse/internal/core/db"
	"github.com/aryanwalia/synapse/internal/core/domain"
)

// SQLWebhookRepository implements domain.WebhookRepository.
type SQLWebhookRepository struct {
	db coreDB.DB
}

func NewSQLWebhookRepository(db coreDB.DB) domain.WebhookRepository {
	return &SQLWebhookRepository{db: db}
}
