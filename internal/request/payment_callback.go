package request

import "github.com/shopspring/decimal"

type TossReturnRequest struct {
	PaymentKey string          `json:"paymentKey"`
	OrderID    string          `json:"orderId"`
	Amount     decimal.Decimal `json:"amount"`
}

type StripeCheckoutSession struct {
	ID                string `json:"id"`
	ClientReferenceID string `json:"client_reference_id"`
	AmountTotal       int64  `json:"amount_total"`
	Currency          string `json:"currency"`
	PaymentStatus     string `json:"payment_status"`
}

type StripeWebhookRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object StripeCheckoutSession `json:"object"`
	} `json:"data"`
}

type AlipayNotifyRequest struct {
	OrderID     string `form:"order_id"`
	TradeNo     string `form:"trade_no"`
	TotalAmount string `form:"total_amount"`
}
