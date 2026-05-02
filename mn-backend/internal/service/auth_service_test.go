package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
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

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     codeRepo.mustGet("user@example.com").Code,
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

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	_, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     codeRepo.mustGet("user@example.com").Code,
	})
	if err != nil {
		t.Fatalf("first register returned error: %v", err)
	}

	_, err = svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     codeRepo.mustGet("user@example.com").Code,
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

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     codeRepo.mustGet("user@example.com").Code,
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
	if !resp.Sent {
		t.Fatalf("expected sent=true, got %#v", resp)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	stored := codeRepo.mustGet("user@example.com")
	if matched := regexp.MustCompile(`^\d{6}$`).MatchString(stored.Code); !matched {
		t.Fatalf("expected stored 6 digit code, got %#v", stored)
	}
	if time.Until(stored.ExpiresAt) < 4*time.Minute {
		t.Fatalf("expected code expiration around 5 minutes, got %s", time.Until(stored.ExpiresAt))
	}

	mailSender := svc.mailSender.(*testMailSender)
	if len(mailSender.messages) != 1 {
		t.Fatalf("expected one email message, got %d", len(mailSender.messages))
	}
	msg := mailSender.messages[0]
	if msg.To != "user@example.com" {
		t.Fatalf("unexpected email recipient: %#v", msg)
	}
	if !strings.Contains(msg.Subject, "邮箱验证码") {
		t.Fatalf("unexpected email subject: %#v", msg)
	}
	if !strings.Contains(msg.Body, stored.Code) {
		t.Fatalf("expected email body to contain code %s, body=%s", stored.Code, msg.Body)
	}
}

func TestAuthService_SendRegisterCodeReplacesOldCode(t *testing.T) {
	svc := newAuthServiceForTest()

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("first send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	first := codeRepo.mustGet("user@example.com")

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("second send register code returned error: %v", err)
	}

	second := codeRepo.mustGet("user@example.com")
	if first.Code == second.Code {
		t.Fatalf("expected resend to replace old code, first=%#v second=%#v", first, second)
	}
}

func TestAuthService_SendRegisterCodeRejectsRegisteredEmail(t *testing.T) {
	svc := newAuthServiceForTest()

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	_, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     codeRepo.mustGet("user@example.com").Code,
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	_, err = svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	})
	if !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestAuthService_RegisterRejectsInvalidOrUsedCode(t *testing.T) {
	svc := newAuthServiceForTest()

	if _, err := svc.SendRegisterCode(context.Background(), request.SendRegisterCodeRequest{
		Email: "user@example.com",
	}); err != nil {
		t.Fatalf("send register code returned error: %v", err)
	}

	codeRepo := svc.registerCodeRepo.(*testRegisterCodeRepository)
	code := codeRepo.mustGet("user@example.com").Code

	_, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     "000000",
	})
	if !errors.Is(err, ErrInvalidRegisterCode) {
		t.Fatalf("expected ErrInvalidRegisterCode for wrong code, got %v", err)
	}

	registerResp, err := svc.Register(context.Background(), request.RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Code:     code,
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if registerResp == nil || registerResp.AccessToken == "" {
		t.Fatalf("expected auth payload, got %#v", registerResp)
	}

	_, err = svc.Register(context.Background(), request.RegisterRequest{
		Email:    "other@example.com",
		Password: "secret123",
		Code:     code,
	})
	if !errors.Is(err, ErrInvalidRegisterCode) {
		t.Fatalf("expected ErrInvalidRegisterCode for used code, got %v", err)
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
		newTestRegisterCodeRepository(),
		jwtManager,
		&testMailSender{},
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
		newTestRegisterCodeRepository(),
		jwtManager,
		&testMailSender{},
	)
}

type testRegisterCodeRepository struct {
	mu    sync.RWMutex
	codes map[string]entity.RegisterCode
}

func newTestRegisterCodeRepository() *testRegisterCodeRepository {
	return &testRegisterCodeRepository{
		codes: make(map[string]entity.RegisterCode),
	}
}

func (r *testRegisterCodeRepository) FindByEmail(_ context.Context, email string) (*entity.RegisterCode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	code, ok := r.codes[email]
	if !ok {
		return nil, nil
	}
	cloned := code
	return &cloned, nil
}

func (r *testRegisterCodeRepository) Save(_ context.Context, code entity.RegisterCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.codes[code.Email] = code
	return nil
}

func (r *testRegisterCodeRepository) Consume(_ context.Context, email, code string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.codes[email]
	if !ok || current.Code != code || !current.UsedAt.IsZero() || !current.ExpiresAt.After(now) {
		return false, nil
	}
	current.UsedAt = now
	r.codes[email] = current
	return true, nil
}

func (r *testRegisterCodeRepository) mustGet(email string) entity.RegisterCode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	code, ok := r.codes[email]
	if !ok {
		panic("register code not found for " + email)
	}
	return code
}

type testMailMessage struct {
	To      string
	Subject string
	Body    string
}

type testMailSender struct {
	messages []testMailMessage
}

func (s *testMailSender) Send(to, subject, body string) error {
	s.messages = append(s.messages, testMailMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}
