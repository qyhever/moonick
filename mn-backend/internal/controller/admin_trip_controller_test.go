package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/repository/mysql"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

func TestAdminTripControllerUpdateBindsFullPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()
	ctx := context.Background()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138200",
		PasswordHash: "hash",
		Nickname:     "后台用户",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseNow := time.Date(2099, 4, 27, 9, 0, 0, 0, time.Local)
	trip, err := tripRepo.Create(ctx, entity.Trip{
		ID:           2200,
		UserID:       user.ID,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  baseNow.Add(4 * time.Hour),
		SeatCount:    3,
		ContactPhone: "13800138200",
		Status:       entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	adminSvc := service.NewAdminService(userRepo, tripRepo, favoriteRepo)
	controller := NewAdminTripController(adminSvc)

	router := gin.New()
	router.PUT("/api/admin/v1/trips/:id", controller.Update)

	body := map[string]any{
		"tripType":          "passenger_post",
		"fromText":          "苏州",
		"toText":            "南京",
		"departureDate":     baseNow.Add(6 * time.Hour).Format(time.DateOnly),
		"departureTime":     baseNow.Add(6 * time.Hour).Format("15:04"),
		"seatCount":         2,
		"priceAmount":       66.6,
		"isPriceNegotiable": false,
		"contactWechat":     "wx-controller",
		"contactPhone":      "13900000001",
		"remark":            "后台更新",
		"status":            entity.TripStatusClosed,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/trips/"+json.Number("2200").String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rec.Body.String())
	}
	if resp.Code != CodeSuccess {
		t.Fatalf("expected success code, got %d body=%s", resp.Code, rec.Body.String())
	}

	got, err := tripRepo.FindByID(ctx, trip.ID)
	if err != nil {
		t.Fatalf("find updated trip: %v", err)
	}
	if got.TripType != "passenger_post" || got.Remark != "后台更新" || got.PriceAmount != 66.6 || got.Status != entity.TripStatusClosed {
		t.Fatalf("unexpected updated trip: %#v", got)
	}
}

func TestAdminTripControllerUpdateSupportsLegacyStatusPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()
	ctx := context.Background()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138202",
		PasswordHash: "hash",
		Nickname:     "后台用户3",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseNow := time.Date(2099, 4, 27, 9, 0, 0, 0, time.Local)
	trip, err := tripRepo.Create(ctx, entity.Trip{
		ID:           2202,
		UserID:       user.ID,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  baseNow.Add(4 * time.Hour),
		SeatCount:    3,
		ContactPhone: "13800138202",
		Status:       entity.TripStatusActive,
	})
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	adminSvc := service.NewAdminService(userRepo, tripRepo, favoriteRepo)
	controller := NewAdminTripController(adminSvc)

	router := gin.New()
	router.PUT("/api/admin/v1/trips/:id", controller.Update)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/trips/2202", bytes.NewBufferString(`{"status":"closed"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rec.Body.String())
	}
	if resp.Code != CodeSuccess {
		t.Fatalf("expected success code, got %d body=%s", resp.Code, rec.Body.String())
	}

	got, err := tripRepo.FindByID(ctx, trip.ID)
	if err != nil {
		t.Fatalf("find updated trip: %v", err)
	}
	if got.Status != entity.TripStatusClosed {
		t.Fatalf("expected legacy status update to succeed, got %#v", got)
	}
}

func TestAdminTripControllerUpdateRejectsIncompleteFullPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := service.NewAdminService(mysql.NewUserRepository(), mysql.NewTripRepository(), mysql.NewFavoriteRepository())
	controller := NewAdminTripController(adminSvc)

	router := gin.New()
	router.PUT("/api/admin/v1/trips/:id", controller.Update)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/trips/1", bytes.NewBufferString(`{"status":"active","tripType":"driver_post"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rec.Body.String())
	}
	if resp.Code != CodeInvalidParam {
		t.Fatalf("expected invalid param code, got %d body=%s", resp.Code, rec.Body.String())
	}
}

func TestAdminTripControllerUpdateMapsServiceValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepo := mysql.NewUserRepository()
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()
	ctx := context.Background()

	user, err := userRepo.Create(ctx, entity.User{
		Phone:        "13800138201",
		PasswordHash: "hash",
		Nickname:     "后台用户2",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseNow := time.Date(2099, 4, 27, 9, 0, 0, 0, time.Local)
	if _, err := tripRepo.Create(ctx, entity.Trip{
		ID:           2201,
		UserID:       user.ID,
		TripType:     "driver_post",
		FromText:     "上海",
		ToText:       "杭州",
		DepartureAt:  baseNow.Add(4 * time.Hour),
		SeatCount:    3,
		ContactPhone: "13800138201",
		Status:       entity.TripStatusActive,
	}); err != nil {
		t.Fatalf("create trip: %v", err)
	}

	adminSvc := service.NewAdminService(userRepo, tripRepo, favoriteRepo)
	controller := NewAdminTripController(adminSvc)

	router := gin.New()
	router.PUT("/api/admin/v1/trips/:id", controller.Update)

	body := map[string]any{
		"tripType":      "driver_post",
		"fromText":      "上海",
		"toText":        "杭州",
		"departureDate": baseNow.Add(6 * time.Hour).Format(time.DateOnly),
		"departureTime": baseNow.Add(6 * time.Hour).Format("15:04"),
		"seatCount":     2,
		"priceAmount":   -1,
		"contactPhone":  "13900000001",
		"status":        entity.TripStatusActive,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/trips/2201", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rec.Body.String())
	}
	if resp.Code != CodeInvalidParam {
		t.Fatalf("expected invalid param code, got %d body=%s", resp.Code, rec.Body.String())
	}
}
