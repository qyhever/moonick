package mysql

import (
	"context"
	"errors"
	"testing"

	"moonick/internal/model/entity"
)

func TestAdminRepository_DoesNotSeedDefaultAdmin(t *testing.T) {
	repo := NewAdminRepository()

	admin, err := repo.FindByUsername(context.Background(), "admin")
	if err != nil {
		t.Fatalf("find default admin returned error: %v", err)
	}
	if admin != nil {
		t.Fatalf("expected no default admin, got %#v", admin)
	}
}

func TestAdminRepository_UpsertSeedAdmin(t *testing.T) {
	db := newRepositoryTestDB(t)

	repo := NewAdminRepositoryWithDB(db)
	if repo.db == nil {
		t.Fatal("expected repository to use database path")
	}

	admin := entity.Admin{
		ID:           1,
		Username:     "admin",
		PasswordHash: "hash-1",
		Name:         "管理员",
		Status:       "active",
	}

	if err := repo.Upsert(context.Background(), admin); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}

	gotByName, err := repo.FindByUsername(context.Background(), "admin")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if gotByName == nil || gotByName.Name != "管理员" || gotByName.PasswordHash != "hash-1" {
		t.Fatalf("unexpected admin by username: %#v", gotByName)
	}

	admin.PasswordHash = "hash-2"
	admin.Name = "超级管理员"
	if err := repo.Upsert(context.Background(), admin); err != nil {
		t.Fatalf("upsert admin second time: %v", err)
	}

	gotByID, err := repo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if gotByID == nil || gotByID.PasswordHash != "hash-2" || gotByID.Name != "超级管理员" {
		t.Fatalf("unexpected admin by id after upsert: %#v", gotByID)
	}
}

func TestAdminRepository_UpsertRenamedAdminInvalidatesOldUsername(t *testing.T) {
	repo := NewAdminRepository()

	if err := repo.Upsert(context.Background(), entity.Admin{
		ID:           1,
		Username:     "admin-old",
		PasswordHash: "hash-1",
		Name:         "管理员",
		Status:       "active",
	}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	if err := repo.Upsert(context.Background(), entity.Admin{
		ID:           1,
		Username:     "admin-new",
		PasswordHash: "hash-2",
		Name:         "新管理员",
		Status:       "active",
	}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	oldAdmin, err := repo.FindByUsername(context.Background(), "admin-old")
	if err != nil {
		t.Fatalf("find old username: %v", err)
	}
	if oldAdmin != nil {
		t.Fatalf("expected old username to be invalidated, got %#v", oldAdmin)
	}

	newAdmin, err := repo.FindByUsername(context.Background(), "admin-new")
	if err != nil {
		t.Fatalf("find new username: %v", err)
	}
	if newAdmin == nil || newAdmin.ID != 1 || newAdmin.PasswordHash != "hash-2" {
		t.Fatalf("expected new username to resolve to updated admin, got %#v", newAdmin)
	}
}

func TestAdminRepository_Create(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewAdminRepositoryWithDB(db)

	admin, err := repo.Create(context.Background(), entity.Admin{
		Username:     "ops-admin",
		PasswordHash: "hash-1",
		Name:         "运营管理员",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if admin == nil || admin.ID == 0 {
		t.Fatalf("expected created admin with id, got %#v", admin)
	}
	if admin.Username != "ops-admin" || admin.Name != "运营管理员" || admin.Status != "active" {
		t.Fatalf("unexpected created admin: %#v", admin)
	}

	got, err := repo.FindByUsername(context.Background(), "ops-admin")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if got == nil || got.ID != admin.ID {
		t.Fatalf("expected persisted admin, got %#v", got)
	}
}

func TestAdminRepository_CreateRejectsDuplicateUsername(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewAdminRepositoryWithDB(db)

	if _, err := repo.Create(context.Background(), entity.Admin{
		Username:     "ops-admin",
		PasswordHash: "hash-1",
		Name:         "运营管理员",
		Status:       "active",
	}); err != nil {
		t.Fatalf("first create admin: %v", err)
	}

	_, err := repo.Create(context.Background(), entity.Admin{
		Username:     "ops-admin",
		PasswordHash: "hash-2",
		Name:         "重复管理员",
		Status:       "active",
	})
	if !errors.Is(err, ErrAdminUsernameAlreadyExists) {
		t.Fatalf("expected ErrAdminUsernameAlreadyExists, got %v", err)
	}
}
