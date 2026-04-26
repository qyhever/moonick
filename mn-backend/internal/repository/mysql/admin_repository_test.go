package mysql

import (
	"context"
	"testing"
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
