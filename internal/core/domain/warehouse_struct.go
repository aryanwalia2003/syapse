package domain

type Warehouse struct {
	ID           string `json:"id"`
	BrandID      string `json:"brand_id"`
	FacilityCode string `json:"facility_code"` // WMS internal code
	Name         string `json:"name"`
	Pincode      string `json:"pincode"`
	City         string `json:"city"`
	State        string `json:"state"`
	IsActive     bool   `json:"is_active"`
}
