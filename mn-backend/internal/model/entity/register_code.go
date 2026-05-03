package entity

import "time"

type RegisterCode struct {
	Email               string
	Code                string
	ExpiresAt           time.Time
	LastSentAt          time.Time
	SendWindowStartedAt time.Time
	SendCountInWindow   int
	UsedAt              time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
