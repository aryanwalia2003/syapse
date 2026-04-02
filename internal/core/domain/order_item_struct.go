package domain

type OrderItem struct {
	ID         string `json:"id"`
	OrderID    string `json:"order_id"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PricePaise int64  `json:"price_paise"`
}

