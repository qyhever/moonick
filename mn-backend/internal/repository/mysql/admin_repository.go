package mysql

import (
	"context"
	"sync"

	"moonick/internal/model/entity"
)

type AdminRepository struct {
	mu             sync.RWMutex
	adminsByID     map[int64]entity.Admin
	adminIDsByName map[string]int64
}

func NewAdminRepository(admins ...entity.Admin) *AdminRepository {
	repo := &AdminRepository{
		adminsByID:     make(map[int64]entity.Admin),
		adminIDsByName: make(map[string]int64),
	}
	for _, admin := range admins {
		repo.adminsByID[admin.ID] = admin
		repo.adminIDsByName[admin.Username] = admin.ID
	}
	return repo
}

func (r *AdminRepository) FindByUsername(_ context.Context, username string) (*entity.Admin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.adminIDsByName[username]
	if !ok {
		return nil, nil
	}

	admin := r.adminsByID[id]
	return cloneAdmin(admin), nil
}

func (r *AdminRepository) FindByID(_ context.Context, id int64) (*entity.Admin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	admin, ok := r.adminsByID[id]
	if !ok {
		return nil, nil
	}
	return cloneAdmin(admin), nil
}

func cloneAdmin(admin entity.Admin) *entity.Admin {
	copied := admin
	return &copied
}
