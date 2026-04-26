package service

import (
	"context"
	"testing"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/repository/mysql"
)

func TestFavoriteService_ToggleFavorite(t *testing.T) {
	svc := newFavoriteServiceForTest()

	first, err := svc.Toggle(context.Background(), 1001, 2001)
	if err != nil {
		t.Fatalf("first toggle returned error: %v", err)
	}
	if !first.Favorited {
		t.Fatalf("expected first toggle to favorite trip, got %#v", first)
	}

	second, err := svc.Toggle(context.Background(), 1001, 2001)
	if err != nil {
		t.Fatalf("second toggle returned error: %v", err)
	}
	if second.Favorited {
		t.Fatalf("expected second toggle to cancel favorite, got %#v", second)
	}
}

func TestFavoriteService_ListFavoritesKeepsMissingTripPlaceholder(t *testing.T) {
	ctx := context.Background()
	favoriteRepo := mysql.NewFavoriteRepository()
	tripRepo := mysql.NewTripRepository()
	if err := favoriteRepo.Create(ctx, 1001, 3001); err != nil {
		t.Fatalf("seed favorite: %v", err)
	}

	svc := NewFavoriteService(favoriteRepo, tripRepo)
	resp, err := svc.ListFavorites(ctx, 1001, request.ListTripRequest{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListFavorites returned error: %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("expected placeholder favorite item, got %#v", resp)
	}
	if !resp.Items[0].Unavailable || resp.Items[0].ID != 3001 {
		t.Fatalf("expected unavailable placeholder, got %#v", resp.Items[0])
	}
}

func TestFavoriteService_ListFavoritesFiltersBeforePagination(t *testing.T) {
	ctx := context.Background()
	favoriteRepo := mysql.NewFavoriteRepository()
	tripRepo := mysql.NewTripRepository()

	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:                3101,
		UserID:            2001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 4, 26, 10, 0, 0, 0, time.Local),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusActive,
		CreatedAt:         time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("seed trip 3101: %v", err)
	}
	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:                3102,
		UserID:            2001,
		TripType:          "passenger_post",
		FromText:          "苏州",
		ToText:            "南京",
		DepartureAt:       time.Date(2026, 4, 26, 11, 0, 0, 0, time.Local),
		SeatCount:         1,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138001",
		Status:            entity.TripStatusClosed,
		CreatedAt:         time.Date(2026, 4, 24, 11, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("seed trip 3102: %v", err)
	}

	for _, tripID := range []int64{3101, 3102, 3999} {
		if err := favoriteRepo.Create(ctx, 1001, tripID); err != nil {
			t.Fatalf("seed favorite %d: %v", tripID, err)
		}
	}

	svc := NewFavoriteService(favoriteRepo, tripRepo)
	resp, err := svc.ListFavorites(ctx, 1001, request.ListTripRequest{
		PageNum:  1,
		PageSize: 1,
		Status:   entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("ListFavorites returned error: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected filtered total 1, got %#v", resp)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID != 3101 {
		t.Fatalf("expected first page to contain only matching active favorite, got %#v", resp.Items)
	}
}

func newFavoriteServiceForTest() *FavoriteService {
	tripRepo := mysql.NewTripRepository()
	if _, err := tripRepo.Create(context.Background(), entity.Trip{
		ID:                2001,
		UserID:            1002,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 4, 25, 10, 0, 0, 0, time.Local),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusActive,
	}); err != nil {
		panic(err)
	}

	return NewFavoriteService(mysql.NewFavoriteRepository(), tripRepo)
}
