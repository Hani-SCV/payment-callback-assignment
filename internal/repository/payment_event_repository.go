package repository

import (
	"context"

	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"gorm.io/gorm"
)

type PaymentEventRepository struct {
	db *gorm.DB
}

func NewPaymentEventRepository(db *gorm.DB) *PaymentEventRepository {
	return &PaymentEventRepository{
		db: db,
	}
}

func (r *PaymentEventRepository) WithTx(tx *gorm.DB) *PaymentEventRepository {
	return &PaymentEventRepository{
		db: tx,
	}
}

func (r *PaymentEventRepository) Create(
	ctx context.Context,
	event *model.PaymentEvent,
) error {
	return r.db.WithContext(ctx).Create(event).Error
}