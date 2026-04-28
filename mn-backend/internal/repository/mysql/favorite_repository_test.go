package mysql

import (
	"context"
	"testing"
	"time"

	"moonick/internal/model/entity"
)

func TestFavoriteRepository_CreateDeleteAndCount(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	if err := repo.Create(ctx, 1001, 2001); err != nil {
		t.Fatalf("create favorite: %v", err)
	}
	if err := repo.Create(ctx, 1001, 2002); err != nil {
		t.Fatalf("create second favorite: %v", err)
	}

	exists, err := repo.Exists(ctx, 1001, 2001)
	if err != nil {
		t.Fatalf("exists favorite: %v", err)
	}
	if !exists {
		t.Fatal("expected favorite to exist")
	}

	total, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count favorites: %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected total favorites: %d", total)
	}

	byUser, err := repo.CountByUser(ctx, 1001)
	if err != nil {
		t.Fatalf("count favorites by user: %v", err)
	}
	if byUser != 2 {
		t.Fatalf("unexpected user favorite count: %d", byUser)
	}

	if err := repo.Delete(ctx, 1001, 2001); err != nil {
		t.Fatalf("delete favorite: %v", err)
	}
	exists, err = repo.Exists(ctx, 1001, 2001)
	if err != nil {
		t.Fatalf("exists after delete: %v", err)
	}
	if exists {
		t.Fatal("expected favorite to be deleted")
	}
}

func TestFavoriteRepository_ListByUser(t *testing.T) {
	db := newRepositoryTestDB(t)
	tripRepo := NewTripRepository(db)
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	tripIDs := make([]int64, 0, 2)
	for idx, trip := range []entity.Trip{
		{
			UserID:       1001,
			TripType:     "driver_post",
			FromText:     "上海",
			ToText:       "杭州",
			DepartureAt:  time.Date(2026, 5, 1, 8, 0, 0, 0, time.Local),
			SeatCount:    3,
			ContactPhone: "13800138000",
			Status:       entity.TripStatusActive,
		},
		{
			UserID:       1002,
			TripType:     "driver_post",
			FromText:     "苏州",
			ToText:       "南京",
			DepartureAt:  time.Date(2026, 5, 2, 9, 0, 0, 0, time.Local),
			SeatCount:    2,
			ContactPhone: "13800138001",
			Status:       entity.TripStatusActive,
		},
	} {
		created, err := tripRepo.Create(ctx, trip)
		if err != nil {
			t.Fatalf("create trip %d: %v", idx, err)
		}
		tripIDs = append(tripIDs, created.ID)
	}

	if err := repo.Create(ctx, 2001, tripIDs[0]); err != nil {
		t.Fatalf("create favorite 1: %v", err)
	}
	if err := repo.Create(ctx, 2001, tripIDs[1]); err != nil {
		t.Fatalf("create favorite 2: %v", err)
	}
	if err := repo.Create(ctx, 2002, tripIDs[1]); err != nil {
		t.Fatalf("create favorite 3: %v", err)
	}

	items, total, err := repo.List(ctx, entity.FavoriteFilter{
		UserID: 2001,
		Offset: 0,
		Limit:  1,
	})
	if err != nil {
		t.Fatalf("list favorites: %v", err)
	}
	if total != 2 || len(items) != 1 {
		t.Fatalf("unexpected favorites list total=%d len=%d", total, len(items))
	}
	if items[0].UserID != 2001 {
		t.Fatalf("unexpected favorite: %#v", items[0])
	}
}

func TestFavoriteRepository_NilDBFallsBackToMemoryEvenIfSharedDBExists(t *testing.T) {
	sharedDB := newRepositoryTestDB(t)
	SetDB(sharedDB)
	defer SetDB(nil)

	repo := NewFavoriteRepository(nil)
	if repo.db != nil {
		t.Fatal("expected nil db argument to force in-memory repository")
	}

	if err := repo.Create(context.Background(), 1001, 2001); err != nil {
		t.Fatalf("create favorite in memory repo: %v", err)
	}

	total, err := repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count favorites in memory repo: %v", err)
	}
	if total != 1 {
		t.Fatalf("unexpected memory favorite count: %d", total)
	}
}
