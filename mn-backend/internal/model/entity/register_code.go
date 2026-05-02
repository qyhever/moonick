package entity

import "time"

type RegisterCode struct {
	Email     string
	Code      string
	ExpiresAt time.Time
	UsedAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
