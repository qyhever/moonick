package entity

import "time"

type User struct {
	ID            int64
	Phone         string
	PasswordHash  string
	Nickname      string
	AvatarURL     string
	Status        string
	DefaultPhone  string
	DefaultWechat string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	Name         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
