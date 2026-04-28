package mysql

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

type AdminRepository struct {
	db             *sql.DB
	mu             sync.RWMutex
	adminsByID     map[int64]entity.Admin
	adminIDsByName map[string]int64
}

func NewAdminRepositoryWithDB(db *sql.DB, admins ...entity.Admin) *AdminRepository {
	if db != nil {
		return &AdminRepository{db: db}
	}

	return newInMemoryAdminRepository(admins...)
}

func NewAdminRepository(admins ...entity.Admin) *AdminRepository {
	return newInMemoryAdminRepository(admins...)
}

func newInMemoryAdminRepository(admins ...entity.Admin) *AdminRepository {
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

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*entity.Admin, error) {
	if r.db != nil {
		return r.findOne(ctx, `SELECT id, username, password_hash, display_name, status, created_at, updated_at
FROM admins
WHERE username = ?`, username)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.adminIDsByName[username]
	if !ok {
		return nil, nil
	}

	admin := r.adminsByID[id]
	return cloneAdmin(admin), nil
}

func (r *AdminRepository) FindByID(ctx context.Context, id int64) (*entity.Admin, error) {
	if r.db != nil {
		return r.findOne(ctx, `SELECT id, username, password_hash, display_name, status, created_at, updated_at
FROM admins
WHERE id = ?`, id)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	admin, ok := r.adminsByID[id]
	if !ok {
		return nil, nil
	}
	return cloneAdmin(admin), nil
}

func (r *AdminRepository) Upsert(ctx context.Context, admin entity.Admin) error {
	if r.db != nil {
		now := time.Now()
		if admin.Status == "" {
			admin.Status = "active"
		}
		if admin.CreatedAt.IsZero() {
			admin.CreatedAt = now
		}
		admin.UpdatedAt = now

		_, err := r.db.ExecContext(ctx, `INSERT INTO admins (id, username, password_hash, display_name, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	username = VALUES(username),
	password_hash = VALUES(password_hash),
	display_name = VALUES(display_name),
	status = VALUES(status),
	updated_at = VALUES(updated_at)`,
			admin.ID,
			admin.Username,
			admin.PasswordHash,
			admin.Name,
			admin.Status,
			admin.CreatedAt,
			admin.UpdatedAt,
		)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if admin.Status == "" {
		admin.Status = "active"
	}
	now := time.Now()
	if admin.CreatedAt.IsZero() {
		admin.CreatedAt = now
	}
	admin.UpdatedAt = now
	if existing, ok := r.adminsByID[admin.ID]; ok && existing.Username != admin.Username {
		delete(r.adminIDsByName, existing.Username)
	}
	r.adminsByID[admin.ID] = admin
	r.adminIDsByName[admin.Username] = admin.ID
	return nil
}

func (r *AdminRepository) findOne(ctx context.Context, query string, arg any) (*entity.Admin, error) {
	var admin entity.Admin
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.Name,
		&admin.Status,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func cloneAdmin(admin entity.Admin) *entity.Admin {
	copied := admin
	return &copied
}
