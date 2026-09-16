package response

type PaymentCallbackResponse struct {
	Result      string `json:"result"`
	PaymentID   string `json:"payment_id"`
	OrderStatus string `json:"order_status"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
