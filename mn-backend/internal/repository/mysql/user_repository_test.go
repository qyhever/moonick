package mysql

import (
	"context"
	"testing"

	"moonick/internal/model/entity"
)

func TestUserRepository_CreateRejectsDuplicatePhone(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.Create(context.Background(), entity.User{
		Phone:        "13800138000",
		PasswordHash: "hash-1",
		Nickname:     "用户8000",
		Status:       "active",
	}); err != nil {
		t.Fatalf("first create returned error: %v", err)
	}

	if _, err := repo.Create(context.Background(), entity.User{
		Phone:        "13800138000",
		PasswordHash: "hash-2",
		Nickname:     "重复用户",
		Status:       "active",
	}); err == nil {
		t.Fatal("expected duplicate phone create to fail")
	}
}
