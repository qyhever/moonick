package service

import (
	"context"
	"testing"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/repository/mysql"
)

func TestAdminService_GetDashboardSummary(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	userA, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138000",
		PasswordHash: "hash-a",
		Nickname:     "用户A",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user a: %v", err)
	}
	userB, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138001",
		PasswordHash: "hash-b",
		Nickname:     "用户B",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user b: %v", err)
	}

	_, err = tripRepo.Create(ctx, entity.Trip{
		ID:                2001,
		UserID:            userA.ID,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Now().Add(2 * time.Hour),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create published trip: %v", err)
	}
	_, err = tripRepo.Create(ctx, entity.Trip{
		ID:                2002,
		UserID:            userB.ID,
		TripType:          "passenger_post",
		FromText:          "苏州",
		ToText:            "南京",
		DepartureAt:       time.Now().Add(-2 * time.Hour),
		SeatCount:         1,
		IsPriceNegotiable: false,
		ContactPhone:      "13800138001",
		Status:            entity.TripStatusExpired,
	})
	if err != nil {
		t.Fatalf("create expired trip: %v", err)
	}
	if err := favoriteRepo.Create(ctx, userA.ID, 2002); err != nil {
		t.Fatalf("create favorite: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	summary, err := svc.GetDashboardSummary(ctx)
	if err != nil {
		t.Fatalf("GetDashboardSummary returned error: %v", err)
	}

	if summary.TotalUsers != 2 {
		t.Fatalf("expected total users 2, got %#v", summary)
	}
	if summary.TotalTrips != 2 || summary.ActiveTrips != 1 || summary.ExpiredTrips != 1 {
		t.Fatalf("unexpected trip counts: %#v", summary)
	}
	if summary.TotalFavorites != 1 {
		t.Fatalf("expected total favorites 1, got %#v", summary)
	}
}

func TestAdminService_UpdateTripRejectsExpiredStatus(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138002",
		PasswordHash: "hash-a",
		Nickname:     "用户C",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:                2010,
		UserID:            user.ID,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Now().Add(2 * time.Hour),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138002",
		Status:            entity.TripStatusActive,
	}); err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	_, err = svc.UpdateTrip(ctx, 2010, request.AdminUpdateTripRequest{Status: entity.TripStatusExpired})
	if err != ErrTripStatusInvalid {
		t.Fatalf("expected ErrTripStatusInvalid, got %v", err)
	}
}

func TestAdminService_UpdateTripRejectsChangingExpiredTrip(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138003",
		PasswordHash: "hash-a",
		Nickname:     "用户D",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:                2020,
		UserID:            user.ID,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Now().Add(-2 * time.Hour),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138003",
		Status:            entity.TripStatusExpired,
	}); err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	_, err = svc.UpdateTrip(ctx, 2020, request.AdminUpdateTripRequest{Status: entity.TripStatusClosed})
	if err != ErrTripStatusInvalid {
		t.Fatalf("expected ErrTripStatusInvalid, got %v", err)
	}
}

func TestAdminService_GetUserDetailCountsAllPublishedTrips(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138004",
		PasswordHash: "hash-a",
		Nickname:     "用户E",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	for idx, status := range []string{entity.TripStatusActive, entity.TripStatusClosed, entity.TripStatusExpired} {
		if _, err := tripRepo.Create(ctx, entity.Trip{
			ID:                int64(2030 + idx),
			UserID:            user.ID,
			TripType:          "driver_post",
			FromText:          "上海",
			ToText:            "杭州",
			DepartureAt:       time.Now().Add(time.Duration(idx) * time.Hour),
			SeatCount:         3,
			IsPriceNegotiable: true,
			ContactPhone:      "13800138004",
			Status:            status,
		}); err != nil {
			t.Fatalf("create trip %d: %v", idx, err)
		}
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	detail, err := svc.GetUserDetail(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserDetail returned error: %v", err)
	}
	if detail.PublishedTripCount != 3 {
		t.Fatalf("expected publishedTripCount=3, got %#v", detail)
	}
}
