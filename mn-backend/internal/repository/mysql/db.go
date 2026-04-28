package mysql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"moonick/internal/model/entity"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrMySQLDSNRequired = errors.New("mysql dsn is required")

	sharedDBMu sync.RWMutex
	sharedDB   *sql.DB
)

const seedAdminUpsertQuery = `INSERT INTO admins (id, username, password_hash, display_name, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	password_hash = VALUES(password_hash),
	display_name = VALUES(display_name),
	status = VALUES(status),
	updated_at = VALUES(updated_at)`

func OpenDB(dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, ErrMySQLDSNRequired
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func SetDB(db *sql.DB) {
	sharedDBMu.Lock()
	defer sharedDBMu.Unlock()

	sharedDB = db
}

func GetDB() *sql.DB {
	sharedDBMu.RLock()
	defer sharedDBMu.RUnlock()

	return sharedDB
}

func SeedAdmin(ctx context.Context, db *sql.DB, admin entity.Admin) error {
	if db == nil {
		return errors.New("mysql db is nil")
	}

	now := time.Now()
	if admin.Status == "" {
		admin.Status = "active"
	}
	if admin.CreatedAt.IsZero() {
		admin.CreatedAt = now
	}
	admin.UpdatedAt = now

	_, err := db.ExecContext(
		ctx,
		seedAdminUpsertQuery,
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
