package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/repository/mysql"
)

func TestTripService_CreateTripRejectsSameRoute(t *testing.T) {
	svc := newTripServiceForTest()

	_, err := svc.CreateTrip(context.Background(), 1001, request.UpsertTripRequest{
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "上海",
		DepartureDate:     "2026-04-25",
		DepartureTime:     "10:00",
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
	})
	if err == nil {
		t.Fatal("expected same route validation error")
	}
	if !strings.Contains(err.Error(), "起点和终点不能相同") {
		t.Fatalf("expected same route error, got %v", err)
	}
}

func TestTripService_CreateTripDefaultsToActiveStatus(t *testing.T) {
	svc := newTripServiceForTest()

	trip, err := svc.CreateTrip(context.Background(), 1001, request.UpsertTripRequest{
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureDate:     "2026-04-26",
		DepartureTime:     "10:00",
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
	})
	if err != nil {
		t.Fatalf("CreateTrip returned error: %v", err)
	}
	if trip.Status != entity.TripStatusActive {
		t.Fatalf("expected active status, got %#v", trip)
	}
}

func TestTripService_CreateTripPersistsTrimmedRemark(t *testing.T) {
	svc := newTripServiceForTest()

	trip, err := svc.CreateTrip(context.Background(), 1001, request.UpsertTripRequest{
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureDate:     "2026-04-26",
		DepartureTime:     "10:00",
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Remark:            "  需要提前联系  ",
	})
	if err != nil {
		t.Fatalf("CreateTrip returned error: %v", err)
	}
	if trip.Remark != "需要提前联系" {
		t.Fatalf("expected trimmed remark, got %#v", trip.Remark)
	}
}

func TestTripService_UpdateTripPersistsTrimmedRemark(t *testing.T) {
	svc, repo := newTripServiceWithRepoForTest()
	ctx := context.Background()

	created, err := repo.Create(ctx, entity.Trip{
		ID:                6001,
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 4, 26, 10, 0, 0, 0, time.Local),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Remark:            "旧备注",
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("seed trip: %v", err)
	}

	updated, err := svc.UpdateTrip(ctx, 1001, created.ID, request.UpsertTripRequest{
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "苏州",
		DepartureDate:     "2026-04-27",
		DepartureTime:     "11:00",
		SeatCount:         2,
		IsPriceNegotiable: false,
		ContactWechat:     "wx-1001",
		Remark:            "  改后的备注 ",
	})
	if err != nil {
		t.Fatalf("UpdateTrip returned error: %v", err)
	}
	if updated.Remark != "改后的备注" {
		t.Fatalf("expected trimmed remark, got %#v", updated.Remark)
	}
}

func TestTripService_ListTripsUsesPageNumAndDefaultPageSize(t *testing.T) {
	svc, repo := newTripServiceWithRepoForTest()
	ctx := context.Background()

	for i := 0; i < 12; i++ {
		if _, err := repo.Create(ctx, entity.Trip{
			ID:                int64(3000 + i),
			UserID:            1001,
			TripType:          "driver_post",
			FromText:          "上海",
			ToText:            "杭州",
			DepartureAt:       time.Date(2026, 5, 10-i, 9, 0, 0, 0, time.Local),
			SeatCount:         3,
			IsPriceNegotiable: true,
			ContactPhone:      "13800138000",
			Status:            entity.TripStatusActive,
			CreatedAt:         time.Date(2026, 4, 24, 10, i, 0, 0, time.Local),
		}); err != nil {
			t.Fatalf("seed trip %d: %v", i, err)
		}
	}

	resp, err := svc.ListTrips(ctx, request.ListTripRequest{PageNum: 1})
	if err != nil {
		t.Fatalf("ListTrips returned error: %v", err)
	}
	if resp.PageNum != 1 || resp.PageSize != 10 {
		t.Fatalf("expected pageNum=1 pageSize=10, got %#v", resp)
	}
	if len(resp.Items) != 10 {
		t.Fatalf("expected 10 items, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != 3011 {
		t.Fatalf("expected latest created trip first, got %#v", resp.Items[0])
	}
}

func TestTripService_ListMyTripsSortsByCreatedAtDesc(t *testing.T) {
	svc, repo := newTripServiceWithRepoForTest()
	ctx := context.Background()

	if _, err := repo.Create(ctx, entity.Trip{
		ID:                4001,
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 5, 1, 9, 0, 0, 0, time.Local),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusActive,
		CreatedAt:         time.Date(2026, 4, 24, 9, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("seed first trip: %v", err)
	}
	if _, err := repo.Create(ctx, entity.Trip{
		ID:                4002,
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "苏州",
		DepartureAt:       time.Date(2026, 4, 26, 9, 0, 0, 0, time.Local),
		SeatCount:         2,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusFull,
		CreatedAt:         time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("seed second trip: %v", err)
	}

	resp, err := svc.ListMyTrips(ctx, 1001, request.ListTripRequest{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListMyTrips returned error: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != 4002 {
		t.Fatalf("expected newer created trip first, got %#v", resp.Items[0])
	}
}

func TestTripService_UpdateTripStatusRejectsExpiredTrip(t *testing.T) {
	svc, repo := newTripServiceWithRepoForTest()
	ctx := context.Background()

	if _, err := repo.Create(ctx, entity.Trip{
		ID:                5001,
		UserID:            1001,
		TripType:          "driver_post",
		FromText:          "上海",
		ToText:            "杭州",
		DepartureAt:       time.Date(2026, 4, 23, 9, 0, 0, 0, time.Local),
		SeatCount:         3,
		IsPriceNegotiable: true,
		ContactPhone:      "13800138000",
		Status:            entity.TripStatusExpired,
		CreatedAt:         time.Date(2026, 4, 20, 9, 0, 0, 0, time.Local),
	}); err != nil {
		t.Fatalf("seed expired trip: %v", err)
	}

	_, err := svc.UpdateTripStatus(ctx, 1001, 5001, entity.TripStatusActive)
	if err != ErrTripStatusInvalid {
		t.Fatalf("expected ErrTripStatusInvalid, got %v", err)
	}
}

func newTripServiceForTest() *TripService {
	svc, _ := newTripServiceWithRepoForTest()
	return svc
}

func newTripServiceWithRepoForTest() (*TripService, *mysql.TripRepository) {
	repo := mysql.NewTripRepository()
	svc := NewTripService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 4, 24, 9, 0, 0, 0, time.Local)
	}
	return svc, repo
}
