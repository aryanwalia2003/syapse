package domain

import (
	"context"
	"time"
)

type WebhookRepository interface {
	SaveRaw(ctx context.Context, webhook *RawWebhook) error
	UpdateStatus(ctx context.Context, correlationID string, status string) error

	// Outbox and Worker operations
	ClaimPending(ctx context.Context, partitionIndex int, limit int) ([]*RawWebhook, error)
	MarkProcessing(ctx context.Context, id string) error
	MarkDone(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
	IncrementRetry(ctx context.Context, id string) error
	RecoverStuck(ctx context.Context, stuckThreshold time.Duration) (int64, error)
}

type RawWebhook struct {
	ID            string    `json:"id" db:"id"`
	CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	Source        string    `json:"source" db:"source"`
	Payload       []byte    `json:"payload" db:"payload"`
	Headers       []byte    `json:"headers" db:"headers"`
	Status        string    `json:"status" db:"status"`
	ReceivedAt    time.Time `json:"received_at" db:"received_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	VendorOrderID   string `json:"vendor_order_id" db:"vendor_order_id"`
	VendorWebhookID string `json:"vendor_webhook_id" db:"vendor_webhook_id"` // Nullable
	RetryCount      int    `json:"retry_count" db:"retry_count"`
	IsDLQ           bool   `json:"is_dlq" db:"is_dlq"`
	PartitionIndex  int    `json:"partition_index" db:"partition_index"`
}
