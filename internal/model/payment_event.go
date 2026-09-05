package model

import "time"

type PaymentEvent struct {
    ID        int64
    EventID   *string
    PaymentID int64
    EventType string
    Payload   []byte
    CreatedAt time.Time
}