package model

import (
	"time"

	"gorm.io/datatypes"
)

type PaymentEvent struct {
	ID        uint           `gorm:"primaryKey"`
	PaymentID uint           `gorm:"not null"`
	EventID   *string        `gorm:"unique"`
	EventType string         `gorm:"not null"`
	Payload   datatypes.JSON `gorm:"not null"`
	CreatedAt time.Time
}