package entity

import "time"

const (
	TripStatusActive  = "active"
	TripStatusFull    = "full"
	TripStatusClosed  = "closed"
	TripStatusExpired = "expired"
)

type Trip struct {
	ID                int64
	UserID            int64
	TripType          string
	FromText          string
	ToText            string
	DepartureAt       time.Time
	SeatCount         int
	PriceAmount       float64
	IsPriceNegotiable bool
	ContactWechat     string
	ContactPhone      string
	Remark            string
	Status            string
	ClosedReason      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Favorite struct {
	UserID    int64
	TripID    int64
	CreatedAt time.Time
}

type TripFilter struct {
	UserID   *int64
	Statuses []string
	TripType string
	Keyword  string
	IDs      []int64
	Offset   int
	Limit    int
}

type FavoriteFilter struct {
	UserID int64
	Offset int
	Limit  int
}
