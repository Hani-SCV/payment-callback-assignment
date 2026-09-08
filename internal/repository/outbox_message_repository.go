package repository

import (
	"context"

	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"gorm.io/gorm"
)

type OutboxMessageRepository struct {
	db *gorm.DB
}

func NewOutboxMessageRepository(db *gorm.DB) *OutboxMessageRepository {
	return &OutboxMessageRepository{
		db: db,
	}
}

func (r *OutboxMessageRepository) WithTx(tx *gorm.DB) *OutboxMessageRepository {
	return &OutboxMessageRepository{
		db: tx,
	}
}

func (r *OutboxMessageRepository) Create(
	ctx context.Context,
	event *model.OutboxMessage,
) error {
	return r.db.WithContext(ctx).Create(event).Error
}

