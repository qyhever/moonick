package service

import (
	"context"
	"regexp"
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
		Email:    "user@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if registerResp == nil || registerResp.AccessToken == "" {
		t.Fatalf("expected access token, got %#v", registerResp)
	}

	loginResp, err := svc.Login(context.Background(), request.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if loginResp == nil || loginResp.User == nil || loginResp.User.Email != "user@example.com" {
		t.Fatalf("expected login user email to be user@example.com, got %#v", loginResp)
	}
}

func TestAuthService_RegisterRejectsDuplicateEmail(t *testing.T) {
	svc := newAuthServiceForTest()

	_, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("first register returned error: %v", err)
	}

	_, err = svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})
	if err == nil {
		t.Fatal("expected duplicate register to fail")
	}
	if err != ErrEmailAlreadyRegistered {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestAuthService_RefreshUserToken(t *testing.T) {
	svc := newAuthServiceForTest()

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
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
	if refreshResp.User == nil || refreshResp.User.Email != "user@example.com" {
		t.Fatalf("expected refreshed user profile, got %#v", refreshResp)
	}
}

func TestAuthService_SendRegisterCode(t *testing.T) {
	svc := newAuthServiceForTest()

	resp, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	})
	if err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected register code payload")
	}
	if matched := regexp.MustCompile(`^\d{6}$`).MatchString(resp.Code); !matched {
		t.Fatalf("expected 6 digit code, got %#v", resp)
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
