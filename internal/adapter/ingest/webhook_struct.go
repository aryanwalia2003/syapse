package ingest

import (
	"time"
)

type WebhookType string

const (
	WebhookTypeWMSOrderCreation WebhookType = "WMS_ORDER_CREATION"
	WebhookTypeDISStatusUpdate  WebhookType = "DIS_STATUS_UPDATE"
)

type WebhookRequest struct {
	Type              WebhookType         `json:"type"`
	ProviderName      string              `json:"provider_name"`
	Payload           []byte              `json:"payload"`
	Header            map[string][]string `json:"header"`
	ReceivedAt        time.Time           `json:"received_at"`
	VendorOrderIDPath string              `json:"vendor_order_id_path"` // gjson path to extract VendorOrderID
}

type IngestResult struct {
	Success       bool   `json:"success"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}
