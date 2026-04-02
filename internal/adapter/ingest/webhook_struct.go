package ingest

import (
	"time"
)

type WebhookRequest struct {
	Source     string              `json:"source"`
	Payload    []byte              `json:"payload"`
	Header     map[string][]string `json:"header"`
	ReceivedAt time.Time           `json:"received_at"`
}

type IngestResult struct {
	Success       bool   `json:"success"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}
