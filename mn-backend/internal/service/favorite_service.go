package service

import (
	"context"
	"strings"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	"moonick/internal/pkg/pagination"
)

type favoriteRepository interface {
	Exists(ctx context.Context, userID, tripID int64) (bool, error)
	Create(ctx context.Context, userID, tripID int64) error
	Delete(ctx context.Context, userID, tripID int64) error
	List(ctx context.Context, filter entity.FavoriteFilter) ([]*entity.Favorite, int, error)
	Count(ctx context.Context) (int, error)
	CountByUser(ctx context.Context, userID int64) (int, error)
}

type favoriteTripRepository interface {
	FindByID(ctx context.Context, id int64) (*entity.Trip, error)
}

type FavoriteService struct {
	favoriteRepo favoriteRepository
	tripRepo     favoriteTripRepository
}

func NewFavoriteService(favoriteRepo favoriteRepository, tripRepo favoriteTripRepository) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favoriteRepo,
		tripRepo:     tripRepo,
	}
}

func (s *FavoriteService) Toggle(ctx context.Context, userID, tripID int64) (*response.ToggleFavoriteResponse, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, normalizeTripRepoError(err)
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}

	exists, err := s.favoriteRepo.Exists(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.favoriteRepo.Delete(ctx, userID, tripID); err != nil {
			return nil, err
		}
		return &response.ToggleFavoriteResponse{Favorited: false}, nil
	}
	if err := s.favoriteRepo.Create(ctx, userID, tripID); err != nil {
		return nil, err
	}
	return &response.ToggleFavoriteResponse{Favorited: true}, nil
}

func (s *FavoriteService) ListFavorites(ctx context.Context, userID int64, req request.ListTripRequest) (*response.TripListResponse, error) {
	params := pagination.Normalize(req.PageNum, req.PageSize)
	favorites, _, err := s.favoriteRepo.List(ctx, entity.FavoriteFilter{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	filtered := make([]*response.TripSummary, 0, len(favorites))
	for _, favorite := range favorites {
		trip, err := s.tripRepo.FindByID(ctx, favorite.TripID)
		if err != nil {
			return nil, normalizeTripRepoError(err)
		}
		if trip == nil {
			if matchUnavailableFavoriteQuery(req) {
				filtered = append(filtered, toUnavailableTripSummary(favorite.TripID))
			}
			continue
		}
		if !matchFavoriteTripQuery(trip, req) {
			continue
		}
		filtered = append(filtered, toTripSummary(trip, true))
	}

	total := len(filtered)
	start := params.Offset()
	if start > total {
		start = total
	}
	end := total
	if start+params.PageSize < end {
		end = start + params.PageSize
	}
	items := filtered[start:end]

	return &response.TripListResponse{
		Items:    items,
		Total:    total,
		PageNum:  params.Page,
		PageSize: params.PageSize,
	}, nil
}

func matchUnavailableFavoriteQuery(req request.ListTripRequest) bool {
	return strings.TrimSpace(req.TripType) == "" &&
		strings.TrimSpace(req.Status) == "" &&
		strings.TrimSpace(req.Keyword) == ""
}

func matchFavoriteTripQuery(trip *entity.Trip, req request.ListTripRequest) bool {
	if trip == nil {
		return false
	}
	if tripType := strings.TrimSpace(req.TripType); tripType != "" && trip.TripType != tripType {
		return false
	}
	if status := strings.TrimSpace(req.Status); status != "" && trip.Status != status {
		return false
	}
	if keyword := strings.ToLower(strings.TrimSpace(req.Keyword)); keyword != "" {
		haystack := strings.ToLower(trip.FromText + " " + trip.ToText)
		if !strings.Contains(haystack, keyword) {
			return false
		}
	}
	return true
}
