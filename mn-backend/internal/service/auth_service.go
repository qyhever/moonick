package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"moonick/htmls"
	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/pkg/password"
	"moonick/internal/pkg/postal"
	"moonick/internal/repository/mysql"
)

var (
	ErrEmailAlreadyRegistered      = errors.New("该邮箱已注册，请直接登录")
	ErrEmailNotRegistered          = errors.New("该邮箱未注册")
	ErrInvalidUserCredentials      = errors.New("邮箱或密码错误")
	ErrInvalidEmail                = errors.New("请输入有效的邮箱地址")
	ErrInvalidRegisterCode         = errors.New("验证码错误或已失效")
	ErrRegisterCodeSendTooFrequent = errors.New("请勿频繁操作")
	ErrInvalidAdminCredentials     = errors.New("账号或密码错误")
	ErrInvalidRefreshToken         = errors.New("refresh token 无效")
	ErrUserNotFound                = errors.New("用户不存在")
	ErrAdminNotFound               = errors.New("管理员不存在")
	ErrStorageNotConfigured        = errors.New("storage not configured")
	ErrEmptyNickname               = errors.New("昵称不能为空")
	ErrEmptyContact                = errors.New("请填写至少一种联系方式")
	ErrAvatarFileRequired          = errors.New("请选择头像文件")
)

const (
	registerCodeTTL              = 5 * time.Minute
	registerCodeResendWindow     = time.Minute
	registerCodeMaxSendPerWindow = 2
)

type tokenManager interface {
	GenerateAccessToken(subject, role string) (string, error)
	GenerateRefreshToken(subject string) (string, error)
	Parse(token string) (*jwtpkg.Claims, error)
}

type authUserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
}

type authAdminRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.Admin, error)
	FindByID(ctx context.Context, id int64) (*entity.Admin, error)
}

type registerCodeRepository interface {
	FindByEmailAndType(ctx context.Context, email, codeType string) (*entity.RegisterCode, error)
	Save(ctx context.Context, code entity.RegisterCode) error
	ConsumeByType(ctx context.Context, email, codeType, code string, now time.Time) (bool, error)
}

type mailSender interface {
	Send(to, subject, body string) error
}

type AuthService struct {
	userRepo         authUserRepository
	adminRepo        authAdminRepository
	registerCodeRepo registerCodeRepository
	tokenManger      tokenManager
	mailSender       mailSender
}

func NewAuthService(
	userRepo authUserRepository,
	adminRepo authAdminRepository,
	registerCodeRepo registerCodeRepository,
	tokenManager tokenManager,
	mailSender mailSender,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		adminRepo:        adminRepo,
		registerCodeRepo: registerCodeRepo,
		tokenManger:      tokenManager,
		mailSender:       mailSender,
	}
}

func (s *AuthService) Register(ctx context.Context, req request.RegisterRequest) (*response.AuthPayload, error) {
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}

	email := strings.TrimSpace(req.Email)
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidUserCredentials
	}
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}
	if s.registerCodeRepo == nil {
		return nil, ErrInvalidRegisterCode
	}

	consumed, err := s.registerCodeRepo.ConsumeByType(
		ctx,
		email,
		request.VerificationCodeTypeRegister,
		strings.TrimSpace(req.Code),
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, ErrInvalidRegisterCode
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, entity.User{
		Email:        email,
		PasswordHash: hash,
		Nickname:     "用户" + emailSuffix(email),
		Status:       "active",
	})
	if err != nil {
		if errors.Is(err, mysql.ErrUserEmailAlreadyExists) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, err
	}

	return s.buildUserAuthPayload(user)
}

func (s *AuthService) SendVerificationCode(ctx context.Context, req request.SendVerificationCodeRequest) (*response.RegisterCodePayload, error) {
	email := strings.TrimSpace(req.Email)
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	codeType := normalizeVerificationCodeType(req.Type)
	if codeType == "" {
		return nil, ErrInvalidRegisterCode
	}
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}
	if s.registerCodeRepo == nil || s.mailSender == nil {
		return nil, ErrInvalidRegisterCode
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	switch codeType {
	case request.VerificationCodeTypeRegister:
		if user != nil {
			return nil, ErrEmailAlreadyRegistered
		}
	case request.VerificationCodeTypeResetPassword:
		if user == nil {
			return nil, ErrEmailNotRegistered
		}
	default:
		return nil, ErrInvalidRegisterCode
	}

	now := time.Now()
	registerCode, err := s.prepareRegisterCodeForSend(ctx, email, codeType, now)
	if err != nil {
		return nil, err
	}
	if err := s.registerCodeRepo.Save(ctx, *registerCode); err != nil {
		return nil, err
	}
	if err := s.mailSender.Send(email, buildVerificationCodeEmailSubject(codeType), buildVerificationCodeEmailBody(registerCode.Code, codeType)); err != nil {
		return nil, err
	}

	return &response.RegisterCodePayload{
		Sent: true,
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req request.ResetPasswordRequest) error {
	if s.userRepo == nil || s.registerCodeRepo == nil {
		return ErrUserNotFound
	}

	email := strings.TrimSpace(req.Email)
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}
	if strings.TrimSpace(req.Password) == "" {
		return ErrInvalidUserCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrEmailNotRegistered
	}

	consumed, err := s.registerCodeRepo.ConsumeByType(
		ctx,
		email,
		request.VerificationCodeTypeResetPassword,
		strings.TrimSpace(req.Code),
		time.Now(),
	)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrInvalidRegisterCode
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return err
	}

	return normalizeUserRepoError(s.userRepo.UpdatePassword(ctx, user.ID, hash))
}

func (s *AuthService) SendRegisterCode(ctx context.Context, req request.SendVerificationCodeRequest) (*response.RegisterCodePayload, error) {
	if strings.TrimSpace(req.Type) == "" {
		req.Type = request.VerificationCodeTypeRegister
	}
	return s.SendVerificationCode(ctx, req)
}

func (s *AuthService) prepareRegisterCodeForSend(ctx context.Context, email, codeType string, now time.Time) (*entity.RegisterCode, error) {
	current, err := s.registerCodeRepo.FindByEmailAndType(ctx, email, codeType)
	if err != nil {
		return nil, err
	}

	if current == nil || !current.ExpiresAt.After(now) {
		code, err := generateVerificationCode(6)
		if err != nil {
			return nil, err
		}

		return &entity.RegisterCode{
			Email:               email,
			Type:                codeType,
			Code:                code,
			ExpiresAt:           now.Add(registerCodeTTL),
			LastSentAt:          now,
			SendWindowStartedAt: now,
			SendCountInWindow:   1,
		}, nil
	}

	next := *current
	if next.SendWindowStartedAt.IsZero() || now.Sub(next.SendWindowStartedAt) >= registerCodeResendWindow {
		next.SendWindowStartedAt = now
		next.SendCountInWindow = 1
	} else {
		if next.SendCountInWindow >= registerCodeMaxSendPerWindow {
			return nil, ErrRegisterCodeSendTooFrequent
		}
		next.SendCountInWindow++
	}
	next.LastSentAt = now
	next.UsedAt = time.Time{}
	return &next, nil
}

func normalizeVerificationCodeType(codeType string) string {
	switch strings.TrimSpace(codeType) {
	case request.VerificationCodeTypeRegister:
		return request.VerificationCodeTypeRegister
	case request.VerificationCodeTypeResetPassword:
		return request.VerificationCodeTypeResetPassword
	default:
		return ""
	}
}

func buildVerificationCodeEmailSubject(codeType string) string {
	if codeType == request.VerificationCodeTypeResetPassword {
		return "重置密码验证码"
	}
	return "邮箱验证码"
}

func buildVerificationCodeEmailBody(code, codeType string) string {
	body := htmls.RenderRegisterCodeTemplateA1(code)
	if codeType != request.VerificationCodeTypeResetPassword {
		return body
	}

	body = strings.ReplaceAll(body, "邮箱验证码", "重置密码验证码")
	body = strings.ReplaceAll(body, "安全登录与注册验证邮件", "重置密码验证邮件")
	body = strings.ReplaceAll(body, "完成注册或登录操作", "完成重置密码操作")
	return body
}

func (s *AuthService) Login(ctx context.Context, req request.LoginRequest) (*response.AuthPayload, error) {
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidUserCredentials
	}
	if err := password.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidUserCredentials
	}

	return s.buildUserAuthPayload(user)
}

func (s *AuthService) RefreshUserToken(ctx context.Context, refreshToken string) (*response.AuthPayload, error) {
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}

	claims, err := s.tokenManger.Parse(strings.TrimSpace(refreshToken))
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if claims == nil || !strings.EqualFold(claims.TokenType, jwtpkg.TokenTypeRefresh) {
		return nil, ErrInvalidRefreshToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return s.buildUserAuthPayload(user)
}

func (s *AuthService) RefreshAdminToken(ctx context.Context, refreshToken string) (*response.AuthPayload, error) {
	if s.adminRepo == nil {
		return nil, ErrAdminNotFound
	}

	claims, err := s.tokenManger.Parse(strings.TrimSpace(refreshToken))
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if claims == nil || !strings.EqualFold(claims.TokenType, jwtpkg.TokenTypeRefresh) {
		return nil, ErrInvalidRefreshToken
	}

	adminID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	return s.buildAdminAuthPayload(admin)
}

func (s *AuthService) AdminLogin(ctx context.Context, req request.AdminLoginRequest) (*response.AuthPayload, error) {
	if s.adminRepo == nil {
		return nil, ErrAdminNotFound
	}

	admin, err := s.adminRepo.FindByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrInvalidAdminCredentials
	}
	if err := password.Compare(admin.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidAdminCredentials
	}

	return s.buildAdminAuthPayload(admin)
}

func (s *AuthService) AdminProfile(ctx context.Context, adminID int64) (*response.AdminProfile, error) {
	if s.adminRepo == nil {
		return nil, ErrAdminNotFound
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	return &response.AdminProfile{
		ID:       admin.ID,
		Username: admin.Username,
		Name:     admin.Name,
		Status:   admin.Status,
	}, nil
}

func (s *AuthService) buildUserAuthPayload(user *entity.User) (*response.AuthPayload, error) {
	if user == nil {
		return nil, ErrUserNotFound
	}

	subject := strconv.FormatInt(user.ID, 10)
	accessToken, err := s.tokenManger.GenerateAccessToken(subject, "user")
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenManger.GenerateRefreshToken(subject)
	if err != nil {
		return nil, err
	}

	return &response.AuthPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserProfile(user),
	}, nil
}

func (s *AuthService) buildAdminAuthPayload(admin *entity.Admin) (*response.AuthPayload, error) {
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	subject := strconv.FormatInt(admin.ID, 10)
	accessToken, err := s.tokenManger.GenerateAccessToken(subject, "admin")
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenManger.GenerateRefreshToken(subject)
	if err != nil {
		return nil, err
	}

	return &response.AuthPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Admin: &response.AdminProfile{
			ID:       admin.ID,
			Username: admin.Username,
			Name:     admin.Name,
			Status:   admin.Status,
		},
	}, nil
}

func toUserProfile(user *entity.User) *response.UserProfile {
	if user == nil {
		return nil
	}

	return &response.UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Phone:         user.Phone,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Status:        user.Status,
		DefaultWechat: user.DefaultWechat,
		DefaultPhone:  user.DefaultPhone,
	}
}

func generateVerificationCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid verification code length")
	}

	digits := make([]byte, length)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}

func isValidEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	return err == nil && strings.EqualFold(strings.TrimSpace(parsed.Address), strings.TrimSpace(email))
}

type PostalMailSender struct{}

func NewPostalMailSender() *PostalMailSender {
	return &PostalMailSender{}
}

func (s *PostalMailSender) Send(to, subject, body string) error {
	return postal.SendMail(to, subject, body)
}

func emailSuffix(email string) string {
	name := email
	if at := strings.Index(email, "@"); at > 0 {
		name = email[:at]
	}
	if len(name) >= 4 {
		return name[len(name)-4:]
	}
	return name
}
