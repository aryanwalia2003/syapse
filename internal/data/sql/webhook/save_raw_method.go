package webhook

import (
	"context"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/pocketbase/pocketbase/tools/security"
)

func (repo *SQLWebhookRepository) SaveRaw(ctx context.Context, webhook *domain.RawWebhook) error {
	sql := `
		INSERT INTO raw_webhook_payloads (id, correlation_id, source, payload, headers)
		VALUES ({:id}, {:correlation_id}, {:source}, {:payload}, {:headers})
	`

	return repo.db.Execute(ctx, sql, map[string]any{
		"id":             security.RandomString(15),
		"correlation_id": webhook.CorrelationID,
		"source":         webhook.Source,
		"payload":        string(webhook.Payload),
		"headers":        string(webhook.Headers),
	})
}

func (repo *SQLWebhookRepository) UpdateStatus(ctx context.Context, correlationID string, status string) error {
	return nil
}
