package domain

import "time"

type Order struct {
	ID              string    `json:"id" db:"id"`
	BrandID         string    `json:"brand_id" db:"brand_id"`
	WarehouseID     string    `json:"warehouse_id" db:"warehouse_id"`
	WMSOrderID      string    `json:"wms_order_id" db:"wms_order_id"`
	DISAWB          string    `json:"dis_awb" db:"dis_awb"` // Nullable initially
	CanonicalStatus string    `json:"canonical_status" db:"canonical_status"`
	CreatedAt       time.Time `json:"created_at" db:"created"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated"`
}
