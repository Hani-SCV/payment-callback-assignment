package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	PaymentStatusPending   = "PENDING"
	PaymentStatusCompleted = "COMPLETED"
	PaymentStatusFailed    = "FAILED"
	PaymentStatusCanceled  = "CANCELED"
)

type Payment struct {
	ID                    uint
	PublicID              string
	OrderID               int64
	Provider              string
	Status                string
	Amount                decimal.Decimal
	Currency              string
	ExternalTransactionID *string
	CancellationReason    *string
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
