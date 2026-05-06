package task

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type tripExpireRepository interface {
	ExpireTripsBefore(ctx context.Context, before time.Time) (int64, error)
}

type TripExpireTask struct {
	tripRepo tripExpireRepository
}

func NewTripExpireTask(tripRepo tripExpireRepository) *TripExpireTask {
	return &TripExpireTask{tripRepo: tripRepo}
}

func (t *TripExpireTask) Run(ctx context.Context) (int64, error) {
	return t.tripRepo.ExpireTripsBefore(ctx, time.Now())
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type SchedulerConfig struct {
	Interval  time.Duration
	NewTicker func(time.Duration) Ticker
	Run       func(context.Context) (int64, error)
}

type Scheduler struct {
	interval  time.Duration
	newTicker func(time.Duration) Ticker
	run       func(context.Context) (int64, error)
}

type stdTicker struct {
	ticker *time.Ticker
}

func (t *stdTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *stdTicker) Stop() {
	t.ticker.Stop()
}

func NewScheduler(cfg SchedulerConfig) *Scheduler {
	interval := cfg.Interval
	if interval <= 0 {
		interval = time.Minute
	}

	newTicker := cfg.NewTicker
	if newTicker == nil {
		newTicker = func(interval time.Duration) Ticker {
			return &stdTicker{ticker: time.NewTicker(interval)}
		}
	}

	return &Scheduler{
		interval:  interval,
		newTicker: newTicker,
		run:       cfg.Run,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	if s == nil || s.run == nil {
		return
	}

	s.runOnce(ctx)

	ticker := s.newTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C():
			s.runOnce(ctx)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	count, err := s.run(ctx)
	if err != nil {
		zap.L().Error("trip expire task failed", zap.Error(err))
		return
	}

	if count <= 0 {
		return
	}

	zap.L().Info("trip expire task completed", zap.Int64("expiredTrips", count))
}
