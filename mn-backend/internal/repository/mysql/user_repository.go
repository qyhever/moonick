package mysql

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"moonick/internal/model/entity"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
)

type UserRepository struct {
	db             *sql.DB
	mu             sync.RWMutex
	nextID         int64
	usersByID      map[int64]entity.User
	userIDsByEmail map[string]int64
}

func NewUserRepository(dbs ...*sql.DB) *UserRepository {
	if len(dbs) > 0 && dbs[0] != nil {
		return &UserRepository{db: dbs[0]}
	}
	if len(dbs) > 0 {
		return &UserRepository{
			nextID:         1000,
			usersByID:      make(map[int64]entity.User),
			userIDsByEmail: make(map[string]int64),
		}
	}

	if db := GetDB(); db != nil {
		return &UserRepository{db: db}
	}

	return &UserRepository{
		nextID:         1000,
		usersByID:      make(map[int64]entity.User),
		userIDsByEmail: make(map[string]int64),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if r.db != nil {
		return r.findOne(ctx, `SELECT id, email, phone, password_hash, nickname, avatar_url, status, default_phone, default_wechat, created_at, updated_at
FROM users
WHERE email = ?`, email)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.userIDsByEmail[email]
	if !ok {
		return nil, nil
	}

	user := r.usersByID[id]
	return cloneUser(user), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	if r.db != nil {
		return r.findOne(ctx, `SELECT id, email, phone, password_hash, nickname, avatar_url, status, default_phone, default_wechat, created_at, updated_at
FROM users
WHERE id = ?`, id)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	return cloneUser(user), nil
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	if r.db != nil {
		result, err := r.db.ExecContext(ctx, `INSERT INTO users (email, phone, password_hash, nickname, avatar_url, status, default_phone, default_wechat)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			user.Email,
			user.Phone,
			user.PasswordHash,
			user.Nickname,
			user.AvatarURL,
			user.Status,
			user.DefaultPhone,
			user.DefaultWechat,
		)
		if isDuplicateKeyError(err) {
			return nil, ErrUserEmailAlreadyExists
		}
		if err != nil {
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		return r.FindByID(ctx, id)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.userIDsByEmail[user.Email]; exists {
		return nil, ErrUserEmailAlreadyExists
	}

	r.nextID++
	now := time.Now()
	user.ID = r.nextID
	user.CreatedAt = now
	user.UpdatedAt = now
	r.usersByID[user.ID] = user
	r.userIDsByEmail[user.Email] = user.ID
	return cloneUser(user), nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, nickname string) error {
	if r.db != nil {
		return r.updateUser(ctx, userID, `UPDATE users SET nickname = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, nickname, userID)
	}

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

func (r *UserRepository) UpdateContact(ctx context.Context, userID int64, defaultWechat, defaultPhone string) error {
	if r.db != nil {
		return r.updateUser(ctx, userID, `UPDATE users SET default_wechat = ?, default_phone = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, defaultWechat, defaultPhone, userID)
	}

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

func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error {
	if r.db != nil {
		return r.updateUser(ctx, userID, `UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, avatarURL, userID)
	}

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

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	if r.db != nil {
		return r.updateUser(ctx, userID, `UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, passwordHash, userID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	r.usersByID[userID] = user
	return nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int, keyword string) ([]*entity.User, int, error) {
	if r.db != nil {
		return r.listFromDB(ctx, offset, limit, keyword)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]entity.User, 0, len(r.usersByID))
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	for _, user := range r.usersByID {
		if normalizedKeyword != "" {
			haystack := strings.ToLower(user.Email + " " + user.Phone + " " + user.Nickname)
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

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	if r.db != nil {
		var total int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.usersByID), nil
}

func (r *UserRepository) findOne(ctx context.Context, query string, arg any) (*entity.User, error) {
	var user entity.User
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Nickname,
		&user.AvatarURL,
		&user.Status,
		&user.DefaultPhone,
		&user.DefaultWechat,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) listFromDB(ctx context.Context, offset, limit int, keyword string) ([]*entity.User, int, error) {
	normalizedKeyword := strings.TrimSpace(keyword)
	countQuery := `SELECT COUNT(*) FROM users`
	listQuery := `SELECT id, email, phone, password_hash, nickname, avatar_url, status, default_phone, default_wechat, created_at, updated_at
FROM users`
	args := make([]any, 0, 5)
	countArgs := make([]any, 0, 3)
	if normalizedKeyword != "" {
		filter := "%" + normalizedKeyword + "%"
		whereClause := ` WHERE email LIKE ? OR phone LIKE ? OR nickname LIKE ?`
		countQuery += whereClause
		listQuery += whereClause
		countArgs = append(countArgs, filter, filter, filter)
		args = append(args, filter, filter, filter)
	}
	listQuery += ` ORDER BY id DESC`
	if limit > 0 {
		listQuery += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	} else if offset > 0 {
		// Keep DB semantics aligned with the in-memory path: no limit still honors offset.
		listQuery += ` LIMIT 18446744073709551615 OFFSET ?`
		args = append(args, offset)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*entity.User, 0)
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.PasswordHash,
			&user.Nickname,
			&user.AvatarURL,
			&user.Status,
			&user.DefaultPhone,
			&user.DefaultWechat,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *UserRepository) updateUser(ctx context.Context, userID int64, query string, args ...any) error {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		exists, err := r.userExistsByID(ctx, userID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrUserNotFound
		}
	}
	return nil
}

func (r *UserRepository) userExistsByID(ctx context.Context, userID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = ?`, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func cloneUser(user entity.User) *entity.User {
	copied := user
	return &copied
}
