package model

import "time"

type OutboxMessage struct {
	ID               int64
	DeduplicationKey string
	EventType        string
	AggregateType    string
	AggregateID      string
	Payload          []byte
	Status           string
	CreatedAt        time.Time
	PublishedAt      *time.Time
}