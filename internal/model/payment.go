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
	Amount     int64  `json:"amount"`
}

type Payment struct {
	ID                    uint       `gorm:"primaryKey"`
	PublicID              string     `gorm:"column:public_id"`
	OrderID               int64      `gorm:"column:order_id"`
	Provider              string     `gorm:"column:provider"`
	Status                string     `gorm:"column:status"`
	Amount                int64      `gorm:"column:amount"`
	Currency              string     `gorm:"column:currency"`
	ExternalTransactionID *string    `gorm:"column:external_transaction_id"`
	CancellationReason    *string    `gorm:"column:cancellation_reason"`
	CompletedAt           *time.Time `gorm:"column:completed_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}