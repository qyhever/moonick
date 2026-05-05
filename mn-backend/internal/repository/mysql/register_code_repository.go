package mysql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"moonick/internal/model/entity"
)

type RegisterCodeRepository struct {
	db    *sql.DB
	mu    sync.RWMutex
	codes map[string]entity.RegisterCode
}

func NewRegisterCodeRepository(dbs ...*sql.DB) *RegisterCodeRepository {
	if len(dbs) > 0 && dbs[0] != nil {
		return &RegisterCodeRepository{db: dbs[0]}
	}
	if len(dbs) > 0 {
		return &RegisterCodeRepository{
			codes: make(map[string]entity.RegisterCode),
		}
	}
	if db := GetDB(); db != nil {
		return &RegisterCodeRepository{db: db}
	}
	return &RegisterCodeRepository{
		codes: make(map[string]entity.RegisterCode),
	}
}

func (r *RegisterCodeRepository) FindByEmailAndType(ctx context.Context, email, codeType string) (*entity.RegisterCode, error) {
	if r.db != nil {
		var (
			item                entity.RegisterCode
			usedAt              sql.NullTime
			lastSentAt          sql.NullTime
			sendWindowStartedAt sql.NullTime
		)
		err := r.db.QueryRowContext(ctx, `SELECT email, type, code, expires_at, last_sent_at, send_window_started_at, send_count_in_window, used_at, created_at, updated_at
FROM register_codes
WHERE email = ? AND type = ?`, email, codeType).Scan(
			&item.Email,
			&item.Type,
			&item.Code,
			&item.ExpiresAt,
			&lastSentAt,
			&sendWindowStartedAt,
			&item.SendCountInWindow,
			&usedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if usedAt.Valid {
			item.UsedAt = usedAt.Time
		}
		if lastSentAt.Valid {
			item.LastSentAt = lastSentAt.Time
		}
		if sendWindowStartedAt.Valid {
			item.SendWindowStartedAt = sendWindowStartedAt.Time
		}
		return &item, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.codes[buildRegisterCodeCacheKey(email, codeType)]
	if !ok {
		return nil, nil
	}
	cloned := item
	return &cloned, nil
}

func (r *RegisterCodeRepository) Save(ctx context.Context, code entity.RegisterCode) error {
	if r.db != nil {
		_, err := r.db.ExecContext(ctx, `INSERT INTO register_codes (email, type, code, expires_at, last_sent_at, send_window_started_at, send_count_in_window, used_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULL)
ON DUPLICATE KEY UPDATE code = VALUES(code), expires_at = VALUES(expires_at), last_sent_at = VALUES(last_sent_at), send_window_started_at = VALUES(send_window_started_at), send_count_in_window = VALUES(send_count_in_window), used_at = NULL, updated_at = CURRENT_TIMESTAMP`,
			code.Email,
			code.Type,
			code.Code,
			code.ExpiresAt,
			nullTimeArg(code.LastSentAt),
			nullTimeArg(code.SendWindowStartedAt),
			code.SendCountInWindow,
		)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	key := buildRegisterCodeCacheKey(code.Email, code.Type)
	current, ok := r.codes[key]
	if ok {
		code.CreatedAt = current.CreatedAt
	} else {
		code.CreatedAt = now
	}
	if code.SendCountInWindow <= 0 {
		code.SendCountInWindow = 1
	}
	code.UsedAt = time.Time{}
	code.UpdatedAt = now
	r.codes[key] = code
	return nil
}

func nullTimeArg(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func (r *RegisterCodeRepository) ConsumeByType(ctx context.Context, email, codeType, code string, now time.Time) (bool, error) {
	if r.db != nil {
		result, err := r.db.ExecContext(ctx, `UPDATE register_codes
SET used_at = ?, updated_at = CURRENT_TIMESTAMP
WHERE email = ? AND type = ? AND code = ? AND used_at IS NULL AND expires_at > ?`,
			now,
			email,
			codeType,
			code,
			now,
		)
		if err != nil {
			return false, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return false, err
		}
		return rows > 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := buildRegisterCodeCacheKey(email, codeType)
	current, ok := r.codes[key]
	if !ok || current.Code != code || !current.UsedAt.IsZero() || !current.ExpiresAt.After(now) {
		return false, nil
	}
	current.UsedAt = now
	current.UpdatedAt = now
	r.codes[key] = current
	return true, nil
}

func buildRegisterCodeCacheKey(email, codeType string) string {
	return email + ":" + codeType
}
