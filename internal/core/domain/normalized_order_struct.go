package domain

import "time"

type NormalizedOrder struct {
	ReferenceCode  string `json:"reference_code"`
	SourceProvider string `json:"source_provider"`

	OrderDate       time.Time            `json:"order_date"`
	CanonicalStatus CanonicalOrderStatus `json:"canonical_status"`

	Financials NormalizedFinancials `json:"financials"`

	PackageWeightGrams int32                `json:"package_weight_grams"`
	Dimensions         NormalizedDimensions `json:"dimensions"`

	PickupAddress   NormalizedAddress `json:"pickup_address"`
	DeliveryAddress NormalizedAddress `json:"delivery_address"`

	Items []NormalizedOrderItem `json:"items"`
}

type NormalizedAddress struct {
	ContactName  string  `json:"contact_name"`
	Phone        string  `json:"phone"`
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 string  `json:"address_line_2"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	PinCode      string  `json:"pin_code"`
	Country      string  `json:"country"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type NormalizedFinancials struct {
	PaymentMode      PaymentMode `json:"payment_mode"`
	TotalAmountPaise int64       `json:"total_amount_paise"`
	Currency         string      `json:"currency"`
}

type NormalizedDimensions struct {
	LengthCm int32 `json:"length_cm"`
	WidthCm  int32 `json:"width_cm"`
	HeightCm int32 `json:"height_cm"`
}

type NormalizedOrderItem struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	PricePaise  int64  `json:"price_paise"`
	WeightGrams int32  `json:"weight_grams"`
}
