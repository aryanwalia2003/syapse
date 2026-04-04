package webhook

import (
	"context"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/pocketbase/pocketbase/tools/security"
)

func (repo *SQLWebhookRepository) SaveRaw(ctx context.Context, webhook *domain.RawWebhook) error {
	id := webhook.ID
	if id == "" {
		id = security.RandomString(15)
	}

	sql := `
		INSERT INTO raw_webhook_payloads
			(id, correlation_id, source, payload, headers, status,
			 vendor_order_id, vendor_webhook_id, retry_count, is_dlq, partition_index, webhook_type)
		VALUES
			({:id}, {:correlation_id}, {:source}, {:payload}, {:headers}, {:status},
			 {:vendor_order_id}, {:vendor_webhook_id}, {:retry_count}, {:is_dlq}, {:partition_index}, {:webhook_type})
	`

	return repo.db.Execute(ctx, sql, map[string]any{
		"id":                id,
		"correlation_id":    webhook.CorrelationID,
		"source":            webhook.Source,
		"payload":           string(webhook.Payload),
		"headers":           string(webhook.Headers),
		"status":            webhook.Status,
		"vendor_order_id":   webhook.VendorOrderID,
		"vendor_webhook_id": webhook.VendorWebhookID,
		"retry_count":       webhook.RetryCount,
		"is_dlq":            webhook.IsDLQ,
		"partition_index":   webhook.PartitionIndex,
		"webhook_type":      webhook.WebhookType,
	})
}

func (repo *SQLWebhookRepository) UpdateStatus(ctx context.Context, correlationID string, status string) error {
	return nil
}
