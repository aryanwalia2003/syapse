package webhook

import (
	"context"
	"encoding/json"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/pocketbase/pocketbase/tools/security"
)

func (repo *SQLWebhookRepository) SaveRaw(ctx context.Context, webhook *domain.RawWebhook) error {
	payloadBytes, _ := json.Marshal(webhook.Payload)
	headerBytes, _ := json.Marshal(webhook.Headers)

	sql := `
		INSERT INTO raw_webhook_payloads (id, correlation_id, source, payload, headers)
		VALUES ({:id}, {:correlation_id}, {:source}, {:payload}, {:headers})
	`

	return repo.db.Execute(ctx, sql, map[string]any{
		"id":             security.RandomString(15),
		"correlation_id": webhook.CorrelationID,
		"source":         webhook.Source,
		"payload":        string(payloadBytes),
		"headers":        string(headerBytes),
	})
}

func (repo *SQLWebhookRepository) UpdateStatus(ctx context.Context, correlationID string, status string) error {
	return nil
}
