package service

import (
	"context"
	"strings"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	"moonick/internal/pkg/pagination"
)

type adminUserRepository interface {
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	List(ctx context.Context, offset, limit int, keyword string) ([]*entity.User, int, error)
	Count(ctx context.Context) (int, error)
}

type adminTripRepository interface {
	FindByID(ctx context.Context, id int64) (*entity.Trip, error)
	List(ctx context.Context, filter entity.TripFilter) ([]*entity.Trip, int, error)
	Update(ctx context.Context, trip entity.Trip) (*entity.Trip, error)
}

type adminFavoriteRepository interface {
	Count(ctx context.Context) (int, error)
	CountByUser(ctx context.Context, userID int64) (int, error)
}

type AdminService struct {
	userRepo     adminUserRepository
	tripRepo     adminTripRepository
	favoriteRepo adminFavoriteRepository
}

func NewAdminService(userRepo adminUserRepository, tripRepo adminTripRepository, favoriteRepo adminFavoriteRepository) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		tripRepo:     tripRepo,
		favoriteRepo: favoriteRepo,
	}
}

func (s *AdminService) GetDashboardSummary(ctx context.Context) (*response.AdminDashboardSummary, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	trips, totalTrips, err := s.tripRepo.List(ctx, entity.TripFilter{})
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}

	activeTrips := 0
	expiredTrips := 0
	for _, trip := range trips {
		switch trip.Status {
		case entity.TripStatusActive:
			activeTrips++
		case entity.TripStatusExpired:
			expiredTrips++
		}
	}

	totalFavorites, err := s.favoriteRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &response.AdminDashboardSummary{
		TotalUsers:     totalUsers,
		TotalTrips:     totalTrips,
		ActiveTrips:    activeTrips,
		ExpiredTrips:   expiredTrips,
		TotalFavorites: totalFavorites,
	}, nil
}

func (s *AdminService) ListTrips(ctx context.Context, req request.ListTripRequest) (*response.TripListResponse, error) {
	params := pagination.Normalize(req.PageNum, req.PageSize)
	filter := entity.TripFilter{
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

func (s *AdminService) GetTripDetail(ctx context.Context, tripID int64) (*response.TripDetail, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	return toTripDetail(trip, false), nil
}

func (s *AdminService) UpdateTrip(ctx context.Context, tripID int64, req request.AdminUpdateTripRequest) (*response.TripDetail, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	if trip.Status == entity.TripStatusExpired {
		return nil, ErrTripStatusInvalid
	}

	status := strings.TrimSpace(req.Status)
	if !isAdminMutableTripStatus(status) {
		return nil, ErrTripStatusInvalid
	}

	trip.Status = status
	updated, err := s.tripRepo.Update(ctx, *trip)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	return toTripDetail(updated, false), nil
}

func (s *AdminService) ListUsers(ctx context.Context, req request.ListUserRequest) (*response.AdminUserListResponse, error) {
	params := pagination.Normalize(req.PageNum, req.PageSize)
	users, total, err := s.userRepo.List(ctx, params.Offset(), params.PageSize, strings.TrimSpace(req.Keyword))
	if err != nil {
		return nil, err
	}

	items := make([]*response.AdminUserSummary, 0, len(users))
	for _, user := range users {
		items = append(items, &response.AdminUserSummary{
			ID:       user.ID,
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Status:   user.Status,
		})
	}

	return &response.AdminUserListResponse{
		Items:    items,
		Total:    total,
		PageNum:  params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (s *AdminService) GetUserDetail(ctx context.Context, userID int64) (*response.AdminUserDetail, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, normalizeUserRepoError(err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	filter := entity.TripFilter{
		UserID: &userID,
	}
	_, publishedTripCount, err := s.tripRepo.List(ctx, filter)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	favoriteCount, err := s.favoriteRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &response.AdminUserDetail{
		ID:                 user.ID,
		Phone:              user.Phone,
		Nickname:           user.Nickname,
		AvatarURL:          user.AvatarURL,
		Status:             user.Status,
		DefaultWechat:      user.DefaultWechat,
		DefaultPhone:       user.DefaultPhone,
		PublishedTripCount: publishedTripCount,
		FavoriteCount:      favoriteCount,
	}, nil
}

func (s *AdminService) ListUserTrips(ctx context.Context, userID int64, req request.ListTripRequest) (*response.TripListResponse, error) {
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

func isAdminMutableTripStatus(status string) bool {
	switch status {
	case entity.TripStatusActive, entity.TripStatusFull, entity.TripStatusClosed:
		return true
	default:
		return false
	}
}
