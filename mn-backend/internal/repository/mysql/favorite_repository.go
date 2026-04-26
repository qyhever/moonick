package mysql

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

type FavoriteRepository struct {
	mu          sync.RWMutex
	favoritesBy map[string]entity.Favorite
}

func NewFavoriteRepository() *FavoriteRepository {
	return &FavoriteRepository{
		favoritesBy: make(map[string]entity.Favorite),
	}
}

func (r *FavoriteRepository) Exists(_ context.Context, userID, tripID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.favoritesBy[favoriteKey(userID, tripID)]
	return ok, nil
}

func (r *FavoriteRepository) Create(_ context.Context, userID, tripID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := favoriteKey(userID, tripID)
	r.favoritesBy[key] = entity.Favorite{
		UserID:    userID,
		TripID:    tripID,
		CreatedAt: time.Now(),
	}
	return nil
}

func (r *FavoriteRepository) Delete(_ context.Context, userID, tripID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.favoritesBy, favoriteKey(userID, tripID))
	return nil
}

func (r *FavoriteRepository) List(_ context.Context, filter entity.FavoriteFilter) ([]*entity.Favorite, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]entity.Favorite, 0, len(r.favoritesBy))
	for _, favorite := range r.favoritesBy {
		if filter.UserID != 0 && favorite.UserID != filter.UserID {
			continue
		}
		items = append(items, favorite)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].TripID > items[j].TripID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := total
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	result := make([]*entity.Favorite, 0, end-start)
	for _, favorite := range items[start:end] {
		copied := favorite
		result = append(result, &copied)
	}
	return result, total, nil
}

func (r *FavoriteRepository) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.favoritesBy), nil
}

func (r *FavoriteRepository) CountByUser(_ context.Context, userID int64) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := 0
	for _, favorite := range r.favoritesBy {
		if favorite.UserID == userID {
			total++
		}
	}
	return total, nil
}

func favoriteKey(userID, tripID int64) string {
	return fmt.Sprintf("%d:%d", userID, tripID)
}
