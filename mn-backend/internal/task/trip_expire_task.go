package task

import (
	"context"
	"time"
)

type tripExpireRepository interface {
	ExpireTripsBefore(ctx context.Context, before time.Time) error
}

type TripExpireTask struct {
	tripRepo tripExpireRepository
}

func NewTripExpireTask(tripRepo tripExpireRepository) *TripExpireTask {
	return &TripExpireTask{tripRepo: tripRepo}
}

func (t *TripExpireTask) Run(ctx context.Context) error {
	return t.tripRepo.ExpireTripsBefore(ctx, time.Now())
}
