package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	"moonick/internal/pkg/pagination"
	"moonick/internal/pkg/timeutil"
	"moonick/internal/repository/mysql"
)

var (
	ErrTripNotFound             = errors.New("行程不存在")
	ErrTripPermissionDenied     = errors.New("无权操作该行程")
	ErrTripInvalidRoute         = errors.New("起点和终点不能相同")
	ErrTripDepartureInPast      = errors.New("出发时间不能早于当前时间")
	ErrTripContactRequired      = errors.New("请填写至少一种联系方式")
	ErrTripTypeRequired         = errors.New("行程类型不能为空")
	ErrTripSeatCountInvalid     = errors.New("座位数必须大于0")
	ErrTripStatusInvalid        = errors.New("行程状态无效")
	ErrTripDepartureDateInvalid = errors.New("出发日期格式错误")
	ErrTripDepartureTimeInvalid = errors.New("出发时间格式错误")
	ErrTripPriceAmountInvalid   = errors.New("价格格式错误")
)

const maxTripPriceAmount = 99999999.99
const tripPricePrecisionTolerance = 1e-9

type tripRepository interface {
	Create(ctx context.Context, trip entity.Trip) (*entity.Trip, error)
	Update(ctx context.Context, trip entity.Trip) (*entity.Trip, error)
	FindByID(ctx context.Context, id int64) (*entity.Trip, error)
	List(ctx context.Context, filter entity.TripFilter) ([]*entity.Trip, int, error)
}

type TripService struct {
	tripRepo tripRepository
	now      func() time.Time
}

func NewTripService(tripRepo tripRepository) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		now:      time.Now,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, userID int64, req request.UpsertTripRequest) (*response.TripDetail, error) {
	now := s.now()
	departureAt, err := s.validateUpsert(req, now)
	if err != nil {
		return nil, err
	}

	trip, err := s.tripRepo.Create(ctx, entity.Trip{
		UserID:            userID,
		TripType:          strings.TrimSpace(req.TripType),
		FromText:          strings.TrimSpace(req.FromText),
		ToText:            strings.TrimSpace(req.ToText),
		DepartureAt:       departureAt,
		SeatCount:         req.SeatCount,
		IsPriceNegotiable: req.IsPriceNegotiable,
		ContactWechat:     strings.TrimSpace(req.ContactWechat),
		ContactPhone:      strings.TrimSpace(req.ContactPhone),
		Status:            entity.TripStatusActive,
	})
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return toTripDetail(trip, false), nil
}

func (s *TripService) UpdateTrip(ctx context.Context, userID, tripID int64, req request.UpsertTripRequest) (*response.TripDetail, error) {
	current, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if current == nil {
		return nil, ErrTripNotFound
	}
	if current.UserID != userID {
		return nil, ErrTripPermissionDenied
	}

	now := s.now()
	departureAt, err := s.validateUpsert(req, now)
	if err != nil {
		return nil, err
	}

	current.TripType = strings.TrimSpace(req.TripType)
	current.FromText = strings.TrimSpace(req.FromText)
	current.ToText = strings.TrimSpace(req.ToText)
	current.DepartureAt = departureAt
	current.SeatCount = req.SeatCount
	current.IsPriceNegotiable = req.IsPriceNegotiable
	current.ContactWechat = strings.TrimSpace(req.ContactWechat)
	current.ContactPhone = strings.TrimSpace(req.ContactPhone)

	updated, err := s.tripRepo.Update(ctx, *current)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return toTripDetail(updated, false), nil
}

func (s *TripService) GetTripDetail(ctx context.Context, tripID int64) (*response.TripDetail, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	return toTripDetail(trip, false), nil
}

func (s *TripService) ListTrips(ctx context.Context, req request.ListTripRequest) (*response.TripListResponse, error) {
	params := pagination.Normalize(req.PageNum, req.PageSize)
	filter := entity.TripFilter{
		Statuses: []string{entity.TripStatusActive, entity.TripStatusFull},
		TripType: strings.TrimSpace(req.TripType),
		Keyword:  strings.TrimSpace(req.Keyword),
		Offset:   params.Offset(),
		Limit:    params.PageSize,
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		filter.Statuses = []string{status}
	}

	items, total, err := s.tripRepo.List(ctx, filter)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return &response.TripListResponse{
		Items:    toTripSummaries(items, false),
		Total:    total,
		PageNum:  params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (s *TripService) ListMyTrips(ctx context.Context, userID int64, req request.ListTripRequest) (*response.TripListResponse, error) {
	params := pagination.Normalize(req.PageNum, req.PageSize)
	filter := entity.TripFilter{
		UserID:   &userID,
		TripType: strings.TrimSpace(req.TripType),
		Keyword:  strings.TrimSpace(req.Keyword),
		Offset:   params.Offset(),
		Limit:    params.PageSize,
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		filter.Statuses = []string{status}
	}

	items, total, err := s.tripRepo.List(ctx, filter)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return &response.TripListResponse{
		Items:    toTripSummaries(items, false),
		Total:    total,
		PageNum:  params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (s *TripService) UpdateTripStatus(ctx context.Context, userID, tripID int64, status string) (*response.TripDetail, error) {
	current, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if current == nil {
		return nil, ErrTripNotFound
	}
	if current.UserID != userID {
		return nil, ErrTripPermissionDenied
	}
	if current.Status == entity.TripStatusExpired {
		return nil, ErrTripStatusInvalid
	}

	status = strings.TrimSpace(status)
	if !isUserMutableTripStatus(status) {
		return nil, ErrTripStatusInvalid
	}

	current.Status = status
	updated, err := s.tripRepo.Update(ctx, *current)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return toTripDetail(updated, false), nil
}

func (s *TripService) validateUpsert(req request.UpsertTripRequest, now time.Time) (time.Time, error) {
	return validateTripFields(
		req.TripType,
		req.FromText,
		req.ToText,
		req.DepartureDate,
		req.DepartureTime,
		req.SeatCount,
		req.ContactWechat,
		req.ContactPhone,
		now,
	)
}

func parseDeparture(req request.UpsertTripRequest, location *time.Location) (time.Time, error) {
	return parseDepartureFields(req.DepartureDate, req.DepartureTime, location)
}

func validateTripFields(
	tripType, fromText, toText, departureDate, departureTime string,
	seatCount int,
	contactWechat, contactPhone string,
	now time.Time,
) (time.Time, error) {
	if strings.TrimSpace(tripType) == "" {
		return time.Time{}, ErrTripTypeRequired
	}
	if strings.TrimSpace(fromText) == "" || strings.TrimSpace(toText) == "" {
		return time.Time{}, ErrTripInvalidRoute
	}
	if strings.TrimSpace(fromText) == strings.TrimSpace(toText) {
		return time.Time{}, ErrTripInvalidRoute
	}
	if seatCount <= 0 {
		return time.Time{}, ErrTripSeatCountInvalid
	}
	departureAt, err := parseDepartureFields(departureDate, departureTime, now.Location())
	if err != nil {
		return time.Time{}, err
	}
	if departureAt.Before(now) {
		return time.Time{}, ErrTripDepartureInPast
	}
	if strings.TrimSpace(contactWechat) == "" && strings.TrimSpace(contactPhone) == "" {
		return time.Time{}, ErrTripContactRequired
	}
	return departureAt, nil
}

func parseDepartureFields(departureDate, departureTime string, location *time.Location) (time.Time, error) {
	departureDateValue, err := timeutil.ParseDepartureDate(strings.TrimSpace(departureDate), location)
	if err != nil {
		return time.Time{}, ErrTripDepartureDateInvalid
	}
	departureAt, err := timeutil.CombineDeparture(departureDateValue, strings.TrimSpace(departureTime), location)
	if err != nil {
		return time.Time{}, ErrTripDepartureTimeInvalid
	}
	return departureAt, nil
}

func validateTripPriceAmount(priceAmount float64) error {
	if math.IsNaN(priceAmount) || math.IsInf(priceAmount, 0) {
		return ErrTripPriceAmountInvalid
	}
	if priceAmount < 0 || priceAmount > maxTripPriceAmount {
		return ErrTripPriceAmountInvalid
	}
	scaled := priceAmount * 100
	if math.Abs(math.Round(scaled)-scaled) > tripPricePrecisionTolerance {
		return ErrTripPriceAmountInvalid
	}
	return nil
}

func normalizeTripRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, mysql.ErrTripNotFound) || errors.Is(err, ErrTripNotFound) {
		return ErrTripNotFound
	}
	return err
}

func toTripSummaries(items []*entity.Trip, favorited bool) []*response.TripSummary {
	result := make([]*response.TripSummary, 0, len(items))
	for _, item := range items {
		result = append(result, toTripSummary(item, favorited))
	}
	return result
}

func toTripSummary(trip *entity.Trip, favorited bool) *response.TripSummary {
	if trip == nil {
		return nil
	}

	return &response.TripSummary{
		ID:                trip.ID,
		UserID:            trip.UserID,
		TripType:          trip.TripType,
		FromText:          trip.FromText,
		ToText:            trip.ToText,
		DepartureDate:     trip.DepartureAt.Format(time.DateOnly),
		DepartureTime:     trip.DepartureAt.Format("15:04"),
		SeatCount:         trip.SeatCount,
		IsPriceNegotiable: trip.IsPriceNegotiable,
		Status:            trip.Status,
		Favorited:         favorited,
	}
}

func toUnavailableTripSummary(tripID int64) *response.TripSummary {
	return &response.TripSummary{
		ID:          tripID,
		Favorited:   true,
		Unavailable: true,
	}
}

func toTripDetail(trip *entity.Trip, favorited bool) *response.TripDetail {
	if trip == nil {
		return nil
	}

	return &response.TripDetail{
		ID:                trip.ID,
		UserID:            trip.UserID,
		TripType:          trip.TripType,
		FromText:          trip.FromText,
		ToText:            trip.ToText,
		DepartureDate:     trip.DepartureAt.Format(time.DateOnly),
		DepartureTime:     trip.DepartureAt.Format("15:04"),
		SeatCount:         trip.SeatCount,
		PriceAmount:       trip.PriceAmount,
		IsPriceNegotiable: trip.IsPriceNegotiable,
		ContactWechat:     trip.ContactWechat,
		ContactPhone:      trip.ContactPhone,
		Remark:            trip.Remark,
		Status:            trip.Status,
		Favorited:         favorited,
		CreatedAt:         trip.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         trip.UpdatedAt.Format(time.RFC3339),
	}
}

func isUserMutableTripStatus(status string) bool {
	switch status {
	case entity.TripStatusActive, entity.TripStatusFull, entity.TripStatusClosed:
		return true
	default:
		return false
	}
}
