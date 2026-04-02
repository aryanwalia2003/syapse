package domain

import "context"

type WebhookRepository interface {
	SaveRaw(ctx context.Context, webhook *RawWebhook) error
	UpdateStatus(ctx context.Context, correlationID string, status string) error
}

type RawWebhook struct {
	CorrelationID string `json:"correlation_id"`
	Source        string `json:"source"`
	Payload       []byte `json:"payload"`
	Headers       []byte `json:"headers"`
	Status        string `json:"status"`
}
