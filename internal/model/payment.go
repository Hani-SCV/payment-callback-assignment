package model

import "time"

const (
	PaymentStatusPending   = "PENDING"
	PaymentStatusCompleted = "COMPLETED"
	PaymentStatusFailed    = "FAILED"
	PaymentStatusCanceled  = "CANCELED"
)

type TossReturnRequest struct {
	PaymentKey string `json:"paymentKey"`
	OrderID    string `json:"orderId"`
	Amount     int    `json:"amount"`
}

type Payment struct {
	ID                    int64
	PublicID              string
	OrderID               int64
	Provider              string
	Status                string
	Amount                string
	Currency              string
	ExternalTransactionID *string
	CancellationReason    *string
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}