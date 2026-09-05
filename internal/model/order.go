package model

import "time"

type Order struct {
	ID                int64
	PublicID          string
	CustomerReference string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}