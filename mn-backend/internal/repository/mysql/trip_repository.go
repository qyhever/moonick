package mysql

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

var ErrTripNotFound = errors.New("trip not found")

type TripRepository struct {
	mu        sync.RWMutex
	nextID    int64
	tripsByID map[int64]entity.Trip
}

func NewTripRepository() *TripRepository {
	return &TripRepository{
		nextID:    2000,
		tripsByID: make(map[int64]entity.Trip),
	}
}

func (r *TripRepository) Create(_ context.Context, trip entity.Trip) (*entity.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if trip.ID == 0 {
		r.nextID++
		trip.ID = r.nextID
	} else if trip.ID > r.nextID {
		r.nextID = trip.ID
	}
	if trip.CreatedAt.IsZero() {
		trip.CreatedAt = now
	}
	trip.UpdatedAt = now
	r.tripsByID[trip.ID] = trip
	return cloneTrip(trip), nil
}

func (r *TripRepository) Update(_ context.Context, trip entity.Trip) (*entity.Trip, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.tripsByID[trip.ID]
	if !ok {
		return nil, ErrTripNotFound
	}

	trip.CreatedAt = current.CreatedAt
	trip.UpdatedAt = time.Now()
	r.tripsByID[trip.ID] = trip
	return cloneTrip(trip), nil
}

func (r *TripRepository) FindByID(_ context.Context, id int64) (*entity.Trip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	trip, ok := r.tripsByID[id]
	if !ok {
		return nil, nil
	}
	return cloneTrip(trip), nil
}

func (r *TripRepository) List(_ context.Context, filter entity.TripFilter) ([]*entity.Trip, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]entity.Trip, 0, len(r.tripsByID))
	for _, trip := range r.tripsByID {
		if !matchTripFilter(trip, filter) {
			continue
		}
		items = append(items, trip)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := total
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	result := make([]*entity.Trip, 0, end-start)
	for _, trip := range items[start:end] {
		result = append(result, cloneTrip(trip))
	}
	return result, total, nil
}

func (r *TripRepository) ExpireTripsBefore(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, trip := range r.tripsByID {
		if trip.Status != entity.TripStatusActive && trip.Status != entity.TripStatusFull {
			continue
		}
		if trip.DepartureAt.Before(before) {
			trip.Status = entity.TripStatusExpired
			trip.UpdatedAt = time.Now()
			r.tripsByID[id] = trip
		}
	}
	return nil
}

func matchTripFilter(trip entity.Trip, filter entity.TripFilter) bool {
	if filter.UserID != nil && trip.UserID != *filter.UserID {
		return false
	}
	if filter.TripType != "" && trip.TripType != filter.TripType {
		return false
	}
	if len(filter.Statuses) > 0 && !containsString(filter.Statuses, trip.Status) {
		return false
	}
	if len(filter.IDs) > 0 && !containsInt64(filter.IDs, trip.ID) {
		return false
	}
	if filter.Keyword != "" {
		keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
		haystack := strings.ToLower(trip.FromText + " " + trip.ToText)
		if !strings.Contains(haystack, keyword) {
			return false
		}
	}
	return true
}

func cloneTrip(trip entity.Trip) *entity.Trip {
	copied := trip
	return &copied
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsInt64(items []int64, target int64) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
