package domain

import (
	"context"
	"time"
)

type WebhookRepository interface {
	SaveRaw(ctx context.Context, webhook *RawWebhook) error
	UpdateStatus(ctx context.Context, correlationID string, status string) error
}

type RawWebhook struct {
	ID            string    `json:"id" db:"id"`
	CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	Source        string    `json:"source" db:"source"`
	Payload       []byte    `json:"payload" db:"payload"`
	Headers       []byte    `json:"headers" db:"headers"`
	Status        string    `json:"status" db:"status"`
	ReceivedAt    time.Time `json:"received_at"`
}
