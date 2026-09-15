package request

import "github.com/shopspring/decimal"

type TossReturnRequest struct {
	PaymentKey string          `json:"paymentKey"`
	OrderID    string          `json:"orderId"`
	Amount     decimal.Decimal `json:"amount"`
}

type StripeWebhookRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object map[string]any `json:"object"`
	} `json:"data"`
}

type AlipayNotifyRequest struct {
	OrderID     string `form:"order_id"`
	TradeNo     string `form:"trade_no"`
	TotalAmount string `form:"total_amount"`
}
