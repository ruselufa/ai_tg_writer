package prodamus_payments

type ProdamusPaymentRequest struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Email       string  `json:"email"`
	Recurring   bool    `json:"recurrent"`
}

type ProdamusPaymentResponse struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Email       string  `json:"email"`
	Recurring   bool    `json:"recurrent"`
}
