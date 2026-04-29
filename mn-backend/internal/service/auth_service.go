package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/pkg/password"
	"moonick/internal/repository/mysql"
)

var (
	ErrPhoneAlreadyRegistered  = errors.New("该手机号已注册，请直接登录")
	ErrInvalidUserCredentials  = errors.New("手机号或密码错误")
	ErrInvalidAdminCredentials = errors.New("账号或密码错误")
	ErrInvalidRefreshToken     = errors.New("refresh token 无效")
	ErrUserNotFound            = errors.New("用户不存在")
	ErrAdminNotFound           = errors.New("管理员不存在")
	ErrStorageNotConfigured    = errors.New("storage not configured")
	ErrEmptyNickname           = errors.New("昵称不能为空")
	ErrEmptyContact            = errors.New("请填写至少一种联系方式")
	ErrAvatarFileRequired      = errors.New("请选择头像文件")
)

type tokenManager interface {
	GenerateAccessToken(subject, role string) (string, error)
	GenerateRefreshToken(subject string) (string, error)
	Parse(token string) (*jwtpkg.Claims, error)
}

type authUserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	Create(ctx context.Context, user entity.User) (*entity.User, error)
}

type authAdminRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.Admin, error)
	FindByID(ctx context.Context, id int64) (*entity.Admin, error)
}

type AuthService struct {
	userRepo    authUserRepository
	adminRepo   authAdminRepository
	tokenManger tokenManager
}

func NewAuthService(userRepo authUserRepository, adminRepo authAdminRepository, tokenManager tokenManager) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		adminRepo:   adminRepo,
		tokenManger: tokenManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req request.RegisterRequest) (*response.AuthPayload, error) {
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}

	phone := strings.TrimSpace(req.Phone)
	if phone == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidUserCredentials
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, entity.User{
		Phone:        phone,
		PasswordHash: hash,
		Nickname:     "用户" + phoneSuffix(phone),
		Status:       "active",
	})
	if err != nil {
		if errors.Is(err, mysql.ErrUserPhoneAlreadyExists) {
			return nil, ErrPhoneAlreadyRegistered
		}
		return nil, err
	}

	return s.buildUserAuthPayload(user)
}

func (s *AuthService) Login(ctx context.Context, req request.LoginRequest) (*response.AuthPayload, error) {
	if s.userRepo == nil {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.FindByPhone(ctx, strings.TrimSpace(req.Phone))
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
		Phone:         user.Phone,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Status:        user.Status,
		DefaultWechat: user.DefaultWechat,
		DefaultPhone:  user.DefaultPhone,
	}
}

func phoneSuffix(phone string) string {
	if len(phone) >= 4 {
		return phone[len(phone)-4:]
	}
	return phone
}
