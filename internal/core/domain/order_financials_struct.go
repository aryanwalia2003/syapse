package domain

type OrderFinancials struct {
	OrderID        string `json:"order_id"`
	PaymentMode    string `json:"payment_mode"` // PREPAID / COD
	CODAmountPaise int64  `json:"cod_amount_paise"`
	Currency       string `json:"currency"`     // Default "INR"
	// Future-proof: Yahan GST, Discount, etc. add ho sakte hain bina Core Order ko chhede
}
