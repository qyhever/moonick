package service

import (
	"context"
	"testing"
	"time"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/pkg/password"
	"moonick/internal/repository/mysql"
)

func TestAuthService_RegisterAndLogin(t *testing.T) {
	svc := newAuthServiceForTest()

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Phone:    "13800138000",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if registerResp == nil || registerResp.AccessToken == "" {
		t.Fatalf("expected access token, got %#v", registerResp)
	}

	loginResp, err := svc.Login(context.Background(), request.LoginRequest{
		Phone:    "13800138000",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if loginResp == nil || loginResp.User == nil || loginResp.User.Phone != "13800138000" {
		t.Fatalf("expected login user phone to be 13800138000, got %#v", loginResp)
	}
}

func TestAuthService_RegisterRejectsDuplicatePhone(t *testing.T) {
	svc := newAuthServiceForTest()

	_, err := svc.Register(context.Background(), request.RegisterRequest{
		Phone:    "13800138000",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("first register returned error: %v", err)
	}

	_, err = svc.Register(context.Background(), request.RegisterRequest{
		Phone:    "13800138000",
		Password: "secret123",
	})
	if err == nil {
		t.Fatal("expected duplicate register to fail")
	}
	if err != ErrPhoneAlreadyRegistered {
		t.Fatalf("expected ErrPhoneAlreadyRegistered, got %v", err)
	}
}

func TestAuthService_RefreshUserToken(t *testing.T) {
	svc := newAuthServiceForTest()

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Phone:    "13800138000",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	refreshResp, err := svc.RefreshUserToken(context.Background(), registerResp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}
	if refreshResp == nil || refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatalf("expected token pair from refresh, got %#v", refreshResp)
	}
	if refreshResp.User == nil || refreshResp.User.Phone != "13800138000" {
		t.Fatalf("expected refreshed user profile, got %#v", refreshResp)
	}
}

func TestAuthService_RefreshAdminToken(t *testing.T) {
	passwordHash, err := password.Hash("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	jwtManager := jwtpkg.NewManager(jwtpkg.Config{
		Secret:          "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	svc := NewAuthService(
		mysql.NewUserRepository(),
		mysql.NewAdminRepository(entity.Admin{
			ID:           1,
			Username:     "root-admin",
			PasswordHash: passwordHash,
			Name:         "Root Admin",
			Status:       "active",
		}),
		jwtManager,
	)

	loginResp, err := svc.AdminLogin(context.Background(), request.AdminLoginRequest{
		Username: "root-admin",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("admin login returned error: %v", err)
	}

	refreshResp, err := svc.RefreshAdminToken(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}
	if refreshResp == nil || refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatalf("expected token pair from refresh, got %#v", refreshResp)
	}
	if refreshResp.Admin == nil || refreshResp.Admin.Username != "root-admin" {
		t.Fatalf("expected refreshed admin profile, got %#v", refreshResp)
	}
}

func newAuthServiceForTest() *AuthService {
	jwtManager := jwtpkg.NewManager(jwtpkg.Config{
		Secret:          "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})

	return NewAuthService(
		mysql.NewUserRepository(),
		mysql.NewAdminRepository(),
		jwtManager,
	)
}
