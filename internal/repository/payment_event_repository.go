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

func (r *PaymentEventRepository) FindByEventID(
	ctx context.Context,
	eventID string,
) (*model.PaymentEvent, error) {
	var event model.PaymentEvent

	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&event).Error
	if err != nil {
		return nil, err
	}

	return &event, nil
}
