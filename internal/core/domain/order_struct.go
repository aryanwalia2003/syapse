package domain

import "time"

type Order struct {
	ID              string    `json:"id"`
	BrandID         string    `json:"brand_id"`
	WarehouseID     string    `json:"warehouse_id"`
	WMSOrderID      string    `json:"wms_order_id"`
	DISAWB          string    `json:"dis_awb"` // Nullable initially
	CanonicalStatus string    `json:"canonical_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
