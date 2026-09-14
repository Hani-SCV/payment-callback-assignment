package repository

import (
	"context"
	"time"

	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

func (r *PaymentRepository) WithTx(tx *gorm.DB) *PaymentRepository {
	return &PaymentRepository{
		db: tx,
	}
}

func (r *PaymentRepository) FindByPublicIDForUpdate(
	ctx context.Context,
	publicID string,
) (*model.Payment, error) {
	payment := &model.Payment{}

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("public_id = ?", publicID).
		Take(payment).Error

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *PaymentRepository) Complete(
	ctx context.Context,
	paymentID uint,
	transactionID string,
) error {
	now := time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Payment{}).
		Where("id = ?", paymentID).
		Updates(map[string]interface{}{
			"status":                  model.PaymentStatusCompleted,
			"external_transaction_id": transactionID,
			"completed_at":            now,
		}).Error
}
