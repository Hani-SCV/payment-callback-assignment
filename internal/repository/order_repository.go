package repository

import (
	"context"

	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) WithTx(tx *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: tx,
	}
}

func (r *OrderRepository) FindByIDForUpdate(
	ctx context.Context,
	id int64,
) (*model.Order, error) {
	order := &model.Order{}

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		Take(order).Error

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepository) MarkAsPaid(
	ctx context.Context,
	orderId int64,
) error {
	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", orderId).
		Update("status", "PAID").Error
}

