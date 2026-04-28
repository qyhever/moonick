package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

type FavoriteRepository struct {
	db          *sql.DB
	mu          sync.RWMutex
	favoritesBy map[string]entity.Favorite
}

func NewFavoriteRepository(dbs ...*sql.DB) *FavoriteRepository {
	if len(dbs) > 0 && dbs[0] != nil {
		return &FavoriteRepository{db: dbs[0]}
	}
	if len(dbs) == 0 {
		if db := GetDB(); db != nil {
			return &FavoriteRepository{db: db}
		}
	}

	return &FavoriteRepository{
		favoritesBy: make(map[string]entity.Favorite),
	}
}

func (r *FavoriteRepository) Exists(ctx context.Context, userID, tripID int64) (bool, error) {
	if r.db != nil {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM trip_favorites WHERE user_id = ? AND trip_id = ?`, userID, tripID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.favoritesBy[favoriteKey(userID, tripID)]
	return ok, nil
}

func (r *FavoriteRepository) Create(ctx context.Context, userID, tripID int64) error {
	if r.db != nil {
		_, err := r.db.ExecContext(ctx, `INSERT INTO trip_favorites (user_id, trip_id, created_at)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE created_at = created_at`, userID, tripID, time.Now())
		return err
	}

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

func (r *FavoriteRepository) Delete(ctx context.Context, userID, tripID int64) error {
	if r.db != nil {
		_, err := r.db.ExecContext(ctx, `DELETE FROM trip_favorites WHERE user_id = ? AND trip_id = ?`, userID, tripID)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.favoritesBy, favoriteKey(userID, tripID))
	return nil
}

func (r *FavoriteRepository) List(ctx context.Context, filter entity.FavoriteFilter) ([]*entity.Favorite, int, error) {
	if r.db != nil {
		return r.listFromDB(ctx, filter)
	}

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

func (r *FavoriteRepository) Count(ctx context.Context) (int, error) {
	if r.db != nil {
		var total int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trip_favorites`).Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.favoritesBy), nil
}

func (r *FavoriteRepository) CountByUser(ctx context.Context, userID int64) (int, error) {
	if r.db != nil {
		var total int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trip_favorites WHERE user_id = ?`, userID).Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}

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

func (r *FavoriteRepository) listFromDB(ctx context.Context, filter entity.FavoriteFilter) ([]*entity.Favorite, int, error) {
	listQuery := `SELECT user_id, trip_id, created_at
FROM trip_favorites
WHERE user_id = ?
ORDER BY created_at DESC, trip_id DESC`
	args := []any{filter.UserID}
	if filter.Limit > 0 {
		listQuery += ` LIMIT ? OFFSET ?`
		args = append(args, filter.Limit, filter.Offset)
	} else if filter.Offset > 0 {
		listQuery += ` LIMIT 18446744073709551615 OFFSET ?`
		args = append(args, filter.Offset)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trip_favorites WHERE user_id = ?`, filter.UserID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*entity.Favorite, 0)
	for rows.Next() {
		var favorite entity.Favorite
		if err := rows.Scan(&favorite.UserID, &favorite.TripID, &favorite.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &favorite)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func favoriteKey(userID, tripID int64) string {
	return fmt.Sprintf("%d:%d", userID, tripID)
}
