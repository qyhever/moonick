package service

import (
	"context"
	"math"
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

func TestAdminService_UpdateTripDetail(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138100",
		PasswordHash: "hash-z",
		Nickname:     "用户Z",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.Local)
	created, err := tripRepo.Create(ctx, entity.Trip{
		ID:                2100,
		UserID:            user.ID,
		TripType:          "driver_post",
		FromText:          "上海南站",
		ToText:            "杭州东站",
		DepartureAt:       now.Add(4 * time.Hour),
		SeatCount:         3,
		PriceAmount:       50,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138100",
		Remark:            "原备注",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	svc.now = func() time.Time { return now }

	detail, err := svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:          "passenger_post",
		FromText:          " 苏州站 ",
		ToText:            " 南京南站 ",
		DepartureDate:     now.Add(6 * time.Hour).Format(time.DateOnly),
		DepartureTime:     now.Add(6 * time.Hour).Format("15:04"),
		SeatCount:         2,
		PriceAmount:       float64Ptr(88.5),
		IsPriceNegotiable: boolPtr(false),
		ContactWechat:     "wx-admin",
		ContactPhone:      " 13900000000 ",
		Remark:            stringPtr(" 需要提前联系 "),
		Status:            entity.TripStatusFull,
	})
	if err != nil {
		t.Fatalf("UpdateTripDetail returned error: %v", err)
	}

	if detail.TripType != "passenger_post" || detail.FromText != "苏州站" || detail.ToText != "南京南站" {
		t.Fatalf("unexpected route fields: %#v", detail)
	}
	if detail.SeatCount != 2 || detail.PriceAmount != 88.5 || detail.IsPriceNegotiable {
		t.Fatalf("unexpected seat/price fields: %#v", detail)
	}
	if detail.ContactWechat != "wx-admin" || detail.ContactPhone != "13900000000" || detail.Remark != "需要提前联系" {
		t.Fatalf("unexpected contact fields: %#v", detail)
	}
	if detail.Status != entity.TripStatusFull {
		t.Fatalf("expected status=%s, got %#v", entity.TripStatusFull, detail)
	}
}

func TestAdminService_UpdateTripDetailPreservesOptionalFieldsWhenOmitted(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138103",
		PasswordHash: "hash-keep",
		Nickname:     "保留用户",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseNow := time.Date(2026, 4, 27, 11, 0, 0, 0, time.Local)
	created, err := tripRepo.Create(ctx, entity.Trip{
		ID:                2103,
		UserID:            user.ID,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       baseNow.Add(3 * time.Hour),
		SeatCount:         3,
		PriceAmount:       66.6,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138103",
		Remark:            "保留备注",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	svc.now = func() time.Time { return baseNow }

	detail, err := svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海虹桥",
		ToText:        "杭州东",
		DepartureDate: baseNow.Add(5 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(5 * time.Hour).Format("15:04"),
		SeatCount:     4,
		ContactPhone:  "13800138103",
		Status:        entity.TripStatusClosed,
	})
	if err != nil {
		t.Fatalf("UpdateTripDetail returned error: %v", err)
	}

	if detail.PriceAmount != 66.6 || !detail.IsPriceNegotiable || detail.Remark != "保留备注" {
		t.Fatalf("expected optional fields preserved, got %#v", detail)
	}
}

func TestAdminService_UpdateTripDetailRejectsExpiredTrip(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138101",
		PasswordHash: "hash-expired",
		Nickname:     "过期用户",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:                2101,
		UserID:            user.ID,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Now().Add(-2 * time.Hour),
		SeatCount:         3,
		PriceAmount:       60,
		IsPriceNegotiable: false,
		ContactPhone:      "13800138101",
		Remark:            "过期行程",
		Status:            entity.TripStatusExpired,
	}); err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	svc.now = func() time.Time { return time.Now() }

	_, err = svc.UpdateTripDetail(ctx, 2101, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "苏州",
		DepartureDate: time.Now().Add(2 * time.Hour).Format(time.DateOnly),
		DepartureTime: time.Now().Add(2 * time.Hour).Format("15:04"),
		SeatCount:     2,
		ContactPhone:  "13800138101",
		Status:        entity.TripStatusClosed,
	})
	if err != ErrTripStatusInvalid {
		t.Fatalf("expected ErrTripStatusInvalid, got %v", err)
	}
}

func TestAdminService_UpdateTripDetailValidationErrors(t *testing.T) {
	ctx := context.Background()
	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138102",
		PasswordHash: "hash-v",
		Nickname:     "校验用户",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseNow := time.Date(2026, 4, 27, 10, 0, 0, 0, time.Local)
	created, err := tripRepo.Create(ctx, entity.Trip{
		ID:           2102,
		UserID:       user.ID,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  baseNow.Add(3 * time.Hour),
		SeatCount:    3,
		ContactPhone: "13800138102",
		Status:       entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	svc := NewAdminService(userRepo, tripRepo, favoriteRepo)
	svc.now = func() time.Time { return baseNow }

	_, err = svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "上海",
		DepartureDate: baseNow.Add(4 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(4 * time.Hour).Format("15:04"),
		SeatCount:     2,
		ContactPhone:  "13800138102",
		Status:        entity.TripStatusActive,
	})
	if err != ErrTripInvalidRoute {
		t.Fatalf("expected ErrTripInvalidRoute, got %v", err)
	}

	_, err = svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "苏州",
		DepartureDate: baseNow.Add(4 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(4 * time.Hour).Format("15:04"),
		SeatCount:     2,
		Status:        entity.TripStatusActive,
	})
	if err != ErrTripContactRequired {
		t.Fatalf("expected ErrTripContactRequired, got %v", err)
	}

	_, err = svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "苏州",
		DepartureDate: baseNow.Add(4 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(4 * time.Hour).Format("15:04"),
		SeatCount:     2,
		ContactPhone:  "13800138102",
		Status:        entity.TripStatusExpired,
	})
	if err != ErrTripStatusInvalid {
		t.Fatalf("expected ErrTripStatusInvalid, got %v", err)
	}

	_, err = svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "苏州",
		DepartureDate: baseNow.Add(4 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(4 * time.Hour).Format("15:04"),
		SeatCount:     2,
		PriceAmount:   float64Ptr(-1),
		ContactPhone:  "13800138102",
		Status:        entity.TripStatusActive,
	})
	if err != ErrTripPriceAmountInvalid {
		t.Fatalf("expected ErrTripPriceAmountInvalid, got %v", err)
	}

	_, err = svc.UpdateTripDetail(ctx, created.ID, request.AdminUpdateTripDetailRequest{
		TripType:      "driver_post",
		FromText:      "上海",
		ToText:        "苏州",
		DepartureDate: baseNow.Add(4 * time.Hour).Format(time.DateOnly),
		DepartureTime: baseNow.Add(4 * time.Hour).Format("15:04"),
		SeatCount:     2,
		PriceAmount:   float64Ptr(math.NaN()),
		ContactPhone:  "13800138102",
		Status:        entity.TripStatusActive,
	})
	if err != ErrTripPriceAmountInvalid {
		t.Fatalf("expected ErrTripPriceAmountInvalid for NaN, got %v", err)
	}
}

func float64Ptr(value float64) *float64 {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
