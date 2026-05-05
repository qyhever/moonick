package router

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"moonick/internal/config"
	"moonick/internal/controller"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/pkg/postal"
)

func TestSetupRouter_RegistersProtectedDomains(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	r := SetupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeNeedLogin)

	adminReq := httptest.NewRequest(http.MethodGet, "/api/admin/v1/auth/me", nil)
	adminRec := httptest.NewRecorder()
	r.ServeHTTP(adminRec, adminReq)
	assertResponseCode(t, adminRec, controller.CodeNeedLogin)
}

func TestSetupRouter_UserRouteRejectsAdminAccessToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	adminToken, err := manager.GenerateAccessToken("1", "admin")
	if err != nil {
		t.Fatalf("generate admin access token: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodePermissionDenied)
}

func TestSetupRouter_UserRouteRejectsRefreshToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	refreshToken, err := manager.GenerateRefreshToken("1")
	if err != nil {
		t.Fatalf("generate user refresh token: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeInvalidToken)
}

func TestSetupRouter_UserRefreshSucceedsWithRefreshToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()
	mailbox := useFakeMailSender(t)

	r := SetupRouter()
	code := sendRegisterCodeAndExtract(t, r, mailbox, "user@example.com", "register")
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"secret123","code":"`+code+`"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	r.ServeHTTP(registerRec, registerReq)

	if registerRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, registerRec.Code, registerRec.Body.String())
	}

	var registerResp struct {
		Code controller.MyCode `json:"code"`
		Data struct {
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerRec.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("unmarshal register response: %v, body=%s", err, registerRec.Body.String())
	}
	if registerResp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", registerResp.Code, registerRec.Body.String())
	}
	if registerResp.Data.RefreshToken == "" {
		t.Fatalf("expected refresh token in register response, body=%s", registerRec.Body.String())
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer "+registerResp.Data.RefreshToken)
	refreshRec := httptest.NewRecorder()
	r.ServeHTTP(refreshRec, refreshReq)

	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, refreshRec.Code, refreshRec.Body.String())
	}

	var refreshResp struct {
		Code controller.MyCode `json:"code"`
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			User         struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(refreshRec.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("unmarshal refresh response: %v, body=%s", err, refreshRec.Body.String())
	}
	if refreshResp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", refreshResp.Code, refreshRec.Body.String())
	}
	if refreshResp.Data.AccessToken == "" || refreshResp.Data.RefreshToken == "" {
		t.Fatalf("expected refresh response tokens, body=%s", refreshRec.Body.String())
	}
	if refreshResp.Data.RefreshToken == registerResp.Data.RefreshToken {
		t.Fatalf("expected rotated refresh token, got same token before=%q after=%q", registerResp.Data.RefreshToken, refreshResp.Data.RefreshToken)
	}
	if refreshResp.Data.User.Email != "user@example.com" {
		t.Fatalf("expected refreshed user email, got %#v", refreshResp.Data.User)
	}
}

func TestSetupRouter_RegisterCodeSendsMailWithoutReturningCode(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()
	mailbox := useFakeMailSender(t)

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/code", bytes.NewBufferString(`{"email":"user@example.com","type":"register"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		Code controller.MyCode `json:"code"`
		Data struct {
			Sent bool `json:"sent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal register code response: %v, body=%s", err, rec.Body.String())
	}
	if resp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", resp.Code, rec.Body.String())
	}
	if !resp.Data.Sent {
		t.Fatalf("expected sent=true, got %#v", resp.Data)
	}
	if len(mailbox.messages) != 1 {
		t.Fatalf("expected one email message, got %d", len(mailbox.messages))
	}
	if extractCodeFromBody(mailbox.messages[0].body) == "" {
		t.Fatalf("expected email body to contain 6 digit code, body=%s", mailbox.messages[0].body)
	}
}

func TestSetupRouter_ResetPasswordRequiresDedicatedVerificationCodeType(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()
	mailbox := useFakeMailSender(t)

	r := SetupRouter()
	registerCode := sendRegisterCodeAndExtract(t, r, mailbox, "user@example.com", "register")
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"secret123","code":"`+registerCode+`"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	r.ServeHTTP(registerRec, registerReq)
	assertResponseCode(t, registerRec, controller.CodeSuccess)

	resetCodeReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/code", bytes.NewBufferString(`{"email":"user@example.com","type":"reset_password"}`))
	resetCodeReq.Header.Set("Content-Type", "application/json")
	resetCodeReq.RemoteAddr = "198.51.100.2:1234"
	resetCodeRec := httptest.NewRecorder()
	r.ServeHTTP(resetCodeRec, resetCodeReq)
	assertResponseCode(t, resetCodeRec, controller.CodeSuccess)

	resetCode := extractCodeFromBody(mailbox.messages[len(mailbox.messages)-1].body)
	if resetCode == "" {
		t.Fatalf("expected reset password email body to contain 6 digit code, body=%s", mailbox.messages[len(mailbox.messages)-1].body)
	}

	invalidResetReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", bytes.NewBufferString(`{"email":"user@example.com","password":"secret456","code":"`+registerCode+`"}`))
	invalidResetReq.Header.Set("Content-Type", "application/json")
	invalidResetRec := httptest.NewRecorder()
	r.ServeHTTP(invalidResetRec, invalidResetReq)
	assertResponseCode(t, invalidResetRec, controller.CodeInvalidParam)

	resetReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", bytes.NewBufferString(`{"email":"user@example.com","password":"secret456","code":"`+resetCode+`"}`))
	resetReq.Header.Set("Content-Type", "application/json")
	resetReq.RemoteAddr = "203.0.113.3:1234"
	resetRec := httptest.NewRecorder()
	r.ServeHTTP(resetRec, resetReq)
	assertResponseCode(t, resetRec, controller.CodeSuccess)

	oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"secret123"}`))
	oldLoginReq.Header.Set("Content-Type", "application/json")
	oldLoginRec := httptest.NewRecorder()
	r.ServeHTTP(oldLoginRec, oldLoginReq)
	assertResponseCode(t, oldLoginRec, controller.CodeInvalidPassword)

	newLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"secret456"}`))
	newLoginReq.Header.Set("Content-Type", "application/json")
	newLoginReq.RemoteAddr = "203.0.113.4:1234"
	newLoginRec := httptest.NewRecorder()
	r.ServeHTTP(newLoginRec, newLoginReq)
	assertResponseCode(t, newLoginRec, controller.CodeSuccess)
}

func TestSetupRouter_AdminRouteRejectsRefreshToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	refreshToken, err := manager.GenerateRefreshToken("1")
	if err != nil {
		t.Fatalf("generate admin refresh token: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeInvalidToken)
}

func TestSetupRouter_AdminRefreshSucceedsWithRefreshToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	config.GlobalConfig.Auth.Admin = config.AdminSeedConfig{
		Username: "root-admin",
		Password: "secret123",
		Name:     "Root Admin",
	}

	r := SetupRouter()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", bytes.NewBufferString(`{"username":"root-admin","password":"secret123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, loginRec.Code, loginRec.Body.String())
	}

	var loginResp struct {
		Code controller.MyCode `json:"code"`
		Data struct {
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("unmarshal admin login response: %v, body=%s", err, loginRec.Body.String())
	}
	if loginResp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", loginResp.Code, loginRec.Body.String())
	}
	if loginResp.Data.RefreshToken == "" {
		t.Fatalf("expected refresh token in admin login response, body=%s", loginRec.Body.String())
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer "+loginResp.Data.RefreshToken)
	refreshRec := httptest.NewRecorder()
	r.ServeHTTP(refreshRec, refreshReq)

	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, refreshRec.Code, refreshRec.Body.String())
	}

	var refreshResp struct {
		Code controller.MyCode `json:"code"`
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			Admin        struct {
				Username string `json:"username"`
			} `json:"admin"`
		} `json:"data"`
	}
	if err := json.Unmarshal(refreshRec.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("unmarshal admin refresh response: %v, body=%s", err, refreshRec.Body.String())
	}
	if refreshResp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", refreshResp.Code, refreshRec.Body.String())
	}
	if refreshResp.Data.AccessToken == "" || refreshResp.Data.RefreshToken == "" {
		t.Fatalf("expected refresh response tokens, body=%s", refreshRec.Body.String())
	}
	if refreshResp.Data.RefreshToken == loginResp.Data.RefreshToken {
		t.Fatalf("expected rotated refresh token, got same token before=%q after=%q", loginResp.Data.RefreshToken, refreshResp.Data.RefreshToken)
	}
	if refreshResp.Data.Admin.Username != "root-admin" {
		t.Fatalf("expected refreshed admin username, got %#v", refreshResp.Data.Admin)
	}
}

func TestSetupRouter_UserProfileUpdateReturnsUserNotExistWhenTokenSubjectMissing(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	userToken, err := manager.GenerateAccessToken("9999", "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewBufferString(`{"nickname":"新昵称"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeUserNotExist)
}

func TestSetupRouter_UploadAvatarReturnsUserNotExistWhenTokenSubjectMissing(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	userToken, err := manager.GenerateAccessToken("9999", "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("avatar")); err != nil {
		t.Fatalf("write avatar body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeUserNotExist)
}

func TestSetupRouter_UploadAvatarOnUsersRouteReturnsUserNotExistWhenTokenSubjectMissing(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	userToken, err := manager.GenerateAccessToken("9999", "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("avatar")); err != nil {
		t.Fatalf("write avatar body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeUserNotExist)
}

func TestSetupRouter_AdminRouteRejectsUserToken(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	manager := testJWTManager()
	userToken, err := manager.GenerateAccessToken("user-1", "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodePermissionDenied)
}

func TestSetupRouter_AdminLoginSucceedsWhenAdminSeedConfigured(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	config.GlobalConfig.Auth.Admin = config.AdminSeedConfig{
		Username: "root-admin",
		Password: "secret123",
		Name:     "Root Admin",
	}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", bytes.NewBufferString(`{"username":"root-admin","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp controller.ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rec.Body.String())
	}
	if resp.Code != controller.CodeSuccess {
		t.Fatalf("expected success code, got %d, body=%s", resp.Code, rec.Body.String())
	}
}

func TestSetupRouter_AdminLoginFailsSafelyWhenAdminSeedMissing(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", bytes.NewBufferString(`{"username":"root-admin","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeInvalidPassword)
}

func TestSetupRouter_ReturnsServerErrorWhenJWTConfigInvalid(t *testing.T) {
	previousConfig := config.GlobalConfig
	t.Cleanup(func() {
		config.GlobalConfig = previousConfig
	})

	config.GlobalConfig = &config.Config{}

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeServerBusy)
}

func TestInitMySQLDBPanicsWhenDSNConfiguredButOpenFails(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			MySQL: config.MySQLConfig{
				Addr:   "127.0.0.1:1",
				User:   "root",
				DBName: "moonick",
			},
		},
	}

	assertPanicsWith(t, "初始化 MySQL 失败", func() {
		_ = initMySQLDB(cfg)
	})
}

func TestNewAdminRepositoryFromConfigPanicsWhenSeedFails(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/moonick")
	if err != nil {
		t.Fatalf("open mysql db handle: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close mysql db handle: %v", err)
	}

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Admin: config.AdminSeedConfig{
				Username: "root-admin",
				Password: "secret123",
				Name:     "Root Admin",
			},
		},
	}

	assertPanicsWith(t, "写入管理员 seed 失败", func() {
		_ = newAdminRepositoryFromConfig(cfg, db)
	})
}

func TestSetupRouter_RegistersSpecAlignedUserRoutes(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	r := SetupRouter()

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/me/trips"},
		{method: http.MethodGet, path: "/api/v1/me/favorites"},
		{method: http.MethodPost, path: "/api/v1/trips/1/favorite"},
		{method: http.MethodPatch, path: "/api/v1/trips/1/status"},
	} {
		req := httptest.NewRequest(route.method, route.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assertResponseCode(t, rec, controller.CodeNeedLogin)
	}
}

func TestSetupRouter_RegistersAdminUserTripsRoute(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/users/1/trips", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeNeedLogin)
}

func useValidJWTConfig(t *testing.T) func() {
	t.Helper()

	previousConfig := config.GlobalConfig
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
		},
		Auth: config.AuthConfig{},
	}

	return func() {
		config.GlobalConfig = previousConfig
	}
}

func testJWTManager() *jwtpkg.Manager {
	return jwtpkg.NewManager(jwtpkg.Config{
		Secret:          config.GlobalConfig.JWT.Secret,
		AccessTokenTTL:  config.GlobalConfig.JWT.AccessTokenTTL,
		RefreshTokenTTL: config.GlobalConfig.JWT.RefreshTokenTTL,
	})
}

type fakeMailbox struct {
	messages []fakeMailMessage
}

type fakeMailMessage struct {
	to      string
	subject string
	body    string
}

func useFakeMailSender(t *testing.T) *fakeMailbox {
	t.Helper()

	mailbox := &fakeMailbox{}
	restore := postal.SetSendMailImplForTest(func(to, subject, body string) error {
		mailbox.messages = append(mailbox.messages, fakeMailMessage{
			to:      to,
			subject: subject,
			body:    body,
		})
		return nil
	})
	t.Cleanup(restore)
	return mailbox
}

func sendRegisterCodeAndExtract(t *testing.T, r http.Handler, mailbox *fakeMailbox, email string, codeType string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/code", bytes.NewBufferString(`{"email":"`+email+`","type":"`+codeType+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assertResponseCode(t, rec, controller.CodeSuccess)

	if len(mailbox.messages) == 0 {
		t.Fatal("expected register code email to be sent")
	}

	code := extractCodeFromBody(mailbox.messages[len(mailbox.messages)-1].body)
	if code == "" {
		t.Fatalf("expected email body to contain 6 digit code, body=%s", mailbox.messages[len(mailbox.messages)-1].body)
	}
	return code
}

func extractCodeFromBody(body string) string {
	if matched := regexp.MustCompile(`>\s*(\d{6})\s*<`).FindStringSubmatch(body); len(matched) == 2 {
		return matched[1]
	}
	return ""
}

func assertResponseCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode controller.MyCode) {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp controller.ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rec.Body.String())
	}
	if resp.Code != expectedCode {
		t.Fatalf("expected response code %d, got %d, body=%s", expectedCode, resp.Code, rec.Body.String())
	}
}

func assertPanicsWith(t *testing.T, expected string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic containing %q", expected)
		}

		if !strings.Contains(toPanicMessage(recovered), expected) {
			t.Fatalf("expected panic containing %q, got %v", expected, recovered)
		}
	}()

	fn()
}

func toPanicMessage(v any) string {
	switch value := v.(type) {
	case error:
		return value.Error()
	case string:
		return value
	default:
		return ""
	}
}
