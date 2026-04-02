package domain

import "time"

type OrderStateLog struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	CorrelationID string    `json:"correlation_id"` // Links to raw_webhooks
	PrevStatus    string    `json:"prev_status"`
	NewStatus     string    `json:"new_status"`
	TriggeredBy   string    `json:"triggered_by"` // WMS, DIS, SYSTEM
	CreatedAt     time.Time `json:"created_at"`
}
