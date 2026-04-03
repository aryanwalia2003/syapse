package domain

import "context"

type WMSNormalizer interface {
	ProviderName() string
	NormalizeInboundOrder(ctx context.Context, rawPayload []byte) (*NormalizedOrder, error)
}

type DISNormalizer interface {
	ProviderName() string
	NormalizeStatusUpdate(ctx context.Context, rawPayload []byte) (*NormalizedStatusUpdate, error)
}

type NormalizedStatusUpdate struct {
	ProviderAWB        string               `json:"provider_awb"`
	SynapseOrderID     string               `json:"synapse_order_id"` //this is used to update the order status in the database
	NewCanonicalStatus CanonicalOrderStatus `json:"new_canonical_status"`
	ProviderRawStatus  string               `json:"provider_raw_status"`
	SourceProvider     string               `json:"source_provider"`
}
