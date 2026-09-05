package database

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

func (tm *TransactionManager) WithTransaction(
	ctx context.Context,
	fn func(tx *gorm.DB) error,
) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}