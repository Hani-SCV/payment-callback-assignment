package model

import "time"

type Order struct {
	ID                int64      `gorm:"primaryKey"`
	PublicID          string    `gorm:"column:public_id"`
	CustomerReference string    `gorm:"column:customer_reference"`
	Status            string    `gorm:"column:status"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}