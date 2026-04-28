package mysql

import (
	"context"
	"testing"
	"time"

	"moonick/internal/model/entity"
)

func TestTripRepository_CreateAndFindByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewTripRepository(db)

	departureAt := time.Date(2026, 5, 1, 8, 30, 0, 0, time.Local)
	created, err := repo.Create(context.Background(), entity.Trip{
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       departureAt,
		SeatCount:         3,
		PriceAmount:       88.5,
		IsPriceNegotiable: true,
		ContactWechat:     "wx-trip",
		ContactPhone:      "13800138000",
		Remark:            "车找人",
		Status:            entity.TripStatusActive,
		ClosedReason:      "",
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("find trip: %v", err)
	}
	if got == nil {
		t.Fatal("expected trip")
	}
	if got.UserID != 1001 || got.PriceAmount != 88.5 || got.Remark != "车找人" {
		t.Fatalf("unexpected trip: %#v", got)
	}
	if !got.DepartureAt.Equal(departureAt) {
		t.Fatalf("unexpected departureAt: %s", got.DepartureAt)
	}
}

func TestTripRepository_UpdateAndList(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewTripRepository(db)
	ctx := context.Background()

	first, err := repo.Create(ctx, entity.Trip{
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海虹桥",
		ToText:            "杭州东",
		DepartureAt:       time.Date(2026, 5, 2, 9, 0, 0, 0, time.Local),
		SeatCount:         3,
		PriceAmount:       90,
		IsPriceNegotiable: false,
		ContactWechat:     "wx-1",
		ContactPhone:      "13800138001",
		Remark:            "早班",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create first trip: %v", err)
	}

	second, err := repo.Create(ctx, entity.Trip{
		UserID:            1002,
		TripType:          "passenger_post",
		FromText:          "苏州",
		ToText:            "南京",
		DepartureAt:       time.Date(2026, 5, 3, 10, 0, 0, 0, time.Local),
		SeatCount:         1,
		PriceAmount:       45,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138002",
		Remark:            "可拼车",
		Status:            entity.TripStatusClosed,
		ClosedReason:      "已约满",
	})
	if err != nil {
		t.Fatalf("create second trip: %v", err)
	}

	first.PriceAmount = 120
	first.FromText = "上海虹桥晚班"
	first.Remark = "改成晚班"
	first.Status = entity.TripStatusFull
	first.ClosedReason = "已满座"
	updated, err := repo.Update(ctx, *first)
	if err != nil {
		t.Fatalf("update trip: %v", err)
	}
	if updated.PriceAmount != 120 || updated.Remark != "改成晚班" || updated.Status != entity.TripStatusFull {
		t.Fatalf("unexpected updated trip: %#v", updated)
	}

	userID := int64(1001)
	items, total, err := repo.List(ctx, entity.TripFilter{
		UserID:   &userID,
		TripType: "driver_post",
		Statuses: []string{entity.TripStatusFull},
		IDs:      []int64{first.ID, second.ID},
		Keyword:  "虹桥晚班",
		Offset:   0,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("list trips: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("unexpected list result total=%d len=%d", total, len(items))
	}
	if items[0].ID != first.ID || items[0].ClosedReason != "已满座" {
		t.Fatalf("unexpected listed trip: %#v", items[0])
	}
}

func TestTripRepository_NilDBFallsBackToMemoryEvenIfSharedDBExists(t *testing.T) {
	sharedDB := newRepositoryTestDB(t)
	SetDB(sharedDB)
	defer SetDB(nil)

	repo := NewTripRepository(nil)
	if repo.db != nil {
		t.Fatal("expected nil db argument to force in-memory repository")
	}

	created, err := repo.Create(context.Background(), entity.Trip{
		UserID:       1001,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  time.Now().Add(2 * time.Hour),
		SeatCount:    3,
		ContactPhone: "13800138000",
		Status:       entity.TripStatusActive,
		PriceAmount:  50,
		Remark:       "memory",
		ClosedReason: "",
	})
	if err != nil {
		t.Fatalf("create trip in memory repo: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("find trip in memory repo: %v", err)
	}
	if got == nil || got.Remark != "memory" {
		t.Fatalf("unexpected memory trip: %#v", got)
	}
}

func TestTripRepository_ExpireTripsBefore(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewTripRepository(db)
	ctx := context.Background()

	expiringTrip, err := repo.Create(ctx, entity.Trip{
		UserID:       1001,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  time.Date(2026, 5, 4, 8, 0, 0, 0, time.Local),
		SeatCount:    3,
		ContactPhone: "13800138000",
		Status:       entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create expiring trip: %v", err)
	}

	preservedTrip, err := repo.Create(ctx, entity.Trip{
		UserID:       1002,
		TripType:     "driver_post",
		FromText:     "苏州",
		ToText:       "南京",
		DepartureAt:  time.Date(2026, 5, 5, 12, 0, 0, 0, time.Local),
		SeatCount:    2,
		ContactPhone: "13800138001",
		Status:       entity.TripStatusFull,
	})
	if err != nil {
		t.Fatalf("create preserved trip: %v", err)
	}

	if err := repo.ExpireTripsBefore(ctx, time.Date(2026, 5, 5, 9, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("expire trips: %v", err)
	}

	expired, err := repo.FindByID(ctx, expiringTrip.ID)
	if err != nil {
		t.Fatalf("find expired trip: %v", err)
	}
	if expired == nil || expired.Status != entity.TripStatusExpired {
		t.Fatalf("expected trip to become expired, got %#v", expired)
	}

	preserved, err := repo.FindByID(ctx, preservedTrip.ID)
	if err != nil {
		t.Fatalf("find preserved trip: %v", err)
	}
	if preserved == nil || preserved.Status != entity.TripStatusFull {
		t.Fatalf("expected later trip to keep full status, got %#v", preserved)
	}
}
