package domain

type OrderRecipient struct {
	OrderID     string `json:"order_id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Pincode     string `json:"pincode"`
	City        string `json:"city"`
	State       string `json:"state"`
	FullAddress string `json:"full_address"`
}
