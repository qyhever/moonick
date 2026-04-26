package mysql

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserPhoneAlreadyExists = errors.New("user phone already exists")
)

type UserRepository struct {
	mu             sync.RWMutex
	nextID         int64
	usersByID      map[int64]entity.User
	userIDsByPhone map[string]int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		nextID:         1000,
		usersByID:      make(map[int64]entity.User),
		userIDsByPhone: make(map[string]int64),
	}
}

func (r *UserRepository) FindByPhone(_ context.Context, phone string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.userIDsByPhone[phone]
	if !ok {
		return nil, nil
	}

	user := r.usersByID[id]
	return cloneUser(user), nil
}

func (r *UserRepository) FindByID(_ context.Context, id int64) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	return cloneUser(user), nil
}

func (r *UserRepository) Create(_ context.Context, user entity.User) (*entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.userIDsByPhone[user.Phone]; exists {
		return nil, ErrUserPhoneAlreadyExists
	}

	r.nextID++
	now := time.Now()
	user.ID = r.nextID
	user.CreatedAt = now
	user.UpdatedAt = now
	r.usersByID[user.ID] = user
	r.userIDsByPhone[user.Phone] = user.ID
	return cloneUser(user), nil
}

func (r *UserRepository) UpdateProfile(_ context.Context, userID int64, nickname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.Nickname = nickname
	user.UpdatedAt = time.Now()
	r.usersByID[userID] = user
	return nil
}

func (r *UserRepository) UpdateContact(_ context.Context, userID int64, defaultWechat, defaultPhone string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.DefaultWechat = defaultWechat
	user.DefaultPhone = defaultPhone
	user.UpdatedAt = time.Now()
	r.usersByID[userID] = user
	return nil
}

func (r *UserRepository) UpdateAvatarURL(_ context.Context, userID int64, avatarURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()
	r.usersByID[userID] = user
	return nil
}

func (r *UserRepository) List(_ context.Context, offset, limit int, keyword string) ([]*entity.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]entity.User, 0, len(r.usersByID))
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	for _, user := range r.usersByID {
		if normalizedKeyword != "" {
			haystack := strings.ToLower(user.Phone + " " + user.Nickname)
			if !strings.Contains(haystack, normalizedKeyword) {
				continue
			}
		}
		items = append(items, user)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID > items[j].ID
	})

	total := len(items)
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	result := make([]*entity.User, 0, end-offset)
	for _, user := range items[offset:end] {
		result = append(result, cloneUser(user))
	}
	return result, total, nil
}

func (r *UserRepository) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.usersByID), nil
}

func cloneUser(user entity.User) *entity.User {
	copied := user
	return &copied
}
