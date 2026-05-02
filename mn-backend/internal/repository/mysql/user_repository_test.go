package mysql

import (
	"context"
	"errors"
	"testing"

	"moonick/internal/model/entity"
)

func TestUserRepository_CreateRejectsDuplicateEmail(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewUserRepository(db)

	if _, err := repo.Create(context.Background(), entity.User{
		Email:        "user@example.com",
		Phone:        "13800138000",
		PasswordHash: "hash-1",
		Nickname:     "用户8000",
		Status:       "active",
	}); err != nil {
		t.Fatalf("first create returned error: %v", err)
	}

	_, err := repo.Create(context.Background(), entity.User{
		Email:        "user@example.com",
		Phone:        "13800138000",
		PasswordHash: "hash-2",
		Nickname:     "重复用户",
		Status:       "active",
	})
	if !errors.Is(err, ErrUserEmailAlreadyExists) {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestUserRepository_DatabaseFlow(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewUserRepository(db)
	if repo.db == nil {
		t.Fatal("expected repository to use database path")
	}

	created, err := repo.Create(context.Background(), entity.User{
		Email:        "db-user@example.com",
		Phone:        "13900139000",
		PasswordHash: "hash-db",
		Nickname:     "数据库用户",
		AvatarURL:    "https://example.com/old.png",
		Status:       "active",
		DefaultPhone: "13900139001",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	byEmail, err := repo.FindByEmail(context.Background(), created.Email)
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if byEmail == nil || byEmail.ID != created.ID {
		t.Fatalf("unexpected user by email: %#v", byEmail)
	}

	byID, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if byID == nil || byID.Phone != created.Phone {
		t.Fatalf("unexpected user by id: %#v", byID)
	}

	if err := repo.UpdateProfile(context.Background(), created.ID, "新昵称"); err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if err := repo.UpdateContact(context.Background(), created.ID, "wx-new", "13900139009"); err != nil {
		t.Fatalf("update contact: %v", err)
	}
	if err := repo.UpdateAvatarURL(context.Background(), created.ID, "https://example.com/new.png"); err != nil {
		t.Fatalf("update avatar: %v", err)
	}

	updated, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("find updated user: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated user")
	}
	if updated.Nickname != "新昵称" || updated.DefaultWechat != "wx-new" || updated.DefaultPhone != "13900139009" || updated.AvatarURL != "https://example.com/new.png" {
		t.Fatalf("unexpected updated user: %#v", updated)
	}

	list, total, err := repo.List(context.Background(), 0, 10, "新昵称")
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("unexpected list result total=%d len=%d", total, len(list))
	}
	if list[0].ID != created.ID {
		t.Fatalf("unexpected listed user: %#v", list[0])
	}

	count, err := repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 1 {
		t.Fatalf("unexpected count: %d", count)
	}
}

func TestUserRepository_NilDBFallsBackToMemoryEvenIfSharedDBExists(t *testing.T) {
	sharedDB := newRepositoryTestDB(t)
	SetDB(sharedDB)
	defer SetDB(nil)

	repo := NewUserRepository(nil)
	if repo.db != nil {
		t.Fatal("expected nil db argument to force in-memory repository")
	}

	created, err := repo.Create(context.Background(), entity.User{
		Email:        "memory-user@example.com",
		Phone:        "13700137000",
		PasswordHash: "hash-memory",
		Nickname:     "内存用户",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("create user in memory repo: %v", err)
	}
	if created == nil || created.ID == 0 {
		t.Fatalf("unexpected created user: %#v", created)
	}

	count, err := repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count users in memory repo: %v", err)
	}
	if count != 1 {
		t.Fatalf("unexpected memory repo count: %d", count)
	}
}

func TestUserRepository_UpdateSameValueDoesNotReturnNotFound(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), entity.User{
		Email:         "same-user@example.com",
		Phone:         "13600136000",
		PasswordHash:  "hash-same",
		Nickname:      "同值用户",
		Status:        "active",
		DefaultWechat: "wx-same",
		DefaultPhone:  "13600136001",
		AvatarURL:     "https://example.com/same.png",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := repo.UpdateProfile(context.Background(), created.ID, created.Nickname); err != nil {
		t.Fatalf("update same nickname should not fail: %v", err)
	}
	if err := repo.UpdateContact(context.Background(), created.ID, created.DefaultWechat, created.DefaultPhone); err != nil {
		t.Fatalf("update same contact should not fail: %v", err)
	}
	if err := repo.UpdateAvatarURL(context.Background(), created.ID, created.AvatarURL); err != nil {
		t.Fatalf("update same avatar should not fail: %v", err)
	}
}

func TestUserRepository_ListWithoutLimitStillHonorsOffset(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewUserRepository(db)

	for _, user := range []entity.User{
		{Email: "user1@example.com", Phone: "13500135001", PasswordHash: "hash-1", Nickname: "用户1", Status: "active"},
		{Email: "user2@example.com", Phone: "13500135002", PasswordHash: "hash-2", Nickname: "用户2", Status: "active"},
		{Email: "user3@example.com", Phone: "13500135003", PasswordHash: "hash-3", Nickname: "用户3", Status: "active"},
	} {
		if _, err := repo.Create(context.Background(), user); err != nil {
			t.Fatalf("create user %s: %v", user.Phone, err)
		}
	}

	list, total, err := repo.List(context.Background(), 1, 0, "")
	if err != nil {
		t.Fatalf("list users without limit: %v", err)
	}
	if total != 3 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 users after offset, got %d", len(list))
	}
	if list[0].Phone != "13500135002" || list[1].Phone != "13500135001" {
		t.Fatalf("unexpected users after offset: %#v", list)
	}
}
