package task

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type stubTripExpireRepository struct {
	count  int64
	err    error
	called atomic.Int64
}

func (s *stubTripExpireRepository) ExpireTripsBefore(_ context.Context, _ time.Time) (int64, error) {
	s.called.Add(1)
	return s.count, s.err
}

type fakeTicker struct {
	ch      chan time.Time
	stopped atomic.Bool
}

func newFakeTicker() *fakeTicker {
	return &fakeTicker{ch: make(chan time.Time, 8)}
}

func (t *fakeTicker) C() <-chan time.Time {
	return t.ch
}

func (t *fakeTicker) Stop() {
	t.stopped.Store(true)
}

func (t *fakeTicker) Tick(at time.Time) {
	t.ch <- at
}

func TestTripExpireTask_RunReturnsExpiredCount(t *testing.T) {
	repo := &stubTripExpireRepository{count: 2}
	task := NewTripExpireTask(repo)

	count, err := task.Run(context.Background())
	if err != nil {
		t.Fatalf("run task: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 expired trips, got %d", count)
	}
	if repo.called.Load() != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.called.Load())
	}
}

func TestScheduler_RunExecutesImmediatelyAndOnTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newFakeTicker()
	var runs atomic.Int64

	scheduler := NewScheduler(SchedulerConfig{
		Interval: time.Minute,
		NewTicker: func(time.Duration) Ticker {
			return ticker
		},
		Run: func(context.Context) (int64, error) {
			return runs.Add(1), nil
		},
	})

	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	waitForRuns(t, &runs, 1)
	ticker.Tick(time.Now())
	waitForRuns(t, &runs, 2)

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancel")
	}
}

func TestScheduler_RunStopsAfterContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := newFakeTicker()

	scheduler := NewScheduler(SchedulerConfig{
		Interval: time.Minute,
		NewTicker: func(time.Duration) Ticker {
			return ticker
		},
		Run: func(context.Context) (int64, error) {
			return 0, nil
		},
	})

	done := make(chan struct{})
	go func() {
		scheduler.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancel")
	}
	if !ticker.stopped.Load() {
		t.Fatal("expected ticker to be stopped on exit")
	}
}

func TestScheduler_RunOnceSkipsInfoLogWhenNoTripExpired(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	previous := zap.L()
	zap.ReplaceGlobals(logger)
	defer zap.ReplaceGlobals(previous)

	scheduler := NewScheduler(SchedulerConfig{
		Run: func(context.Context) (int64, error) {
			return 0, nil
		},
	})

	scheduler.runOnce(context.Background())

	if logs.Len() != 0 {
		t.Fatalf("expected no logs when no trip expired, got %d", logs.Len())
	}
}

func TestScheduler_RunOnceWritesInfoLogWhenTripExpired(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	previous := zap.L()
	zap.ReplaceGlobals(logger)
	defer zap.ReplaceGlobals(previous)

	scheduler := NewScheduler(SchedulerConfig{
		Run: func(context.Context) (int64, error) {
			return 3, nil
		},
	})

	scheduler.runOnce(context.Background())

	if logs.Len() != 1 {
		t.Fatalf("expected one info log, got %d", logs.Len())
	}

	entry := logs.All()[0]
	if entry.Message != "trip expire task completed" {
		t.Fatalf("unexpected log message: %s", entry.Message)
	}

	if got := entry.ContextMap()["expiredTrips"]; got != int64(3) {
		t.Fatalf("expected expiredTrips=3, got %v", got)
	}
}

func waitForRuns(t *testing.T, runs *atomic.Int64, want int64) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if runs.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected at least %d runs, got %d", want, runs.Load())
}
