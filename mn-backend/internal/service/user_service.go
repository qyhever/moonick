package service

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
	"moonick/internal/model/response"
	"moonick/internal/repository/mysql"
)

type avatarUploader interface {
	UploadAvatar(ctx context.Context, userID int64, file *multipart.FileHeader) (string, error)
}

type userProfileRepository interface {
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID int64, nickname string) error
	UpdateContact(ctx context.Context, userID int64, defaultWechat, defaultPhone string) error
	UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error
}

type UserService struct {
	userRepo userProfileRepository
	uploader avatarUploader
}

func NewUserService(userRepo userProfileRepository, uploader avatarUploader) *UserService {
	return &UserService{
		userRepo: userRepo,
		uploader: uploader,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*response.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return toUserProfile(user), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req request.UpdateProfileRequest) error {
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		return ErrEmptyNickname
	}
	return normalizeUserRepoError(s.userRepo.UpdateProfile(ctx, userID, nickname))
}

func (s *UserService) UpdateContact(ctx context.Context, userID int64, req request.UpdateContactRequest) error {
	if strings.TrimSpace(req.DefaultWechat) == "" && strings.TrimSpace(req.DefaultPhone) == "" {
		return ErrEmptyContact
	}
	return normalizeUserRepoError(s.userRepo.UpdateContact(ctx, userID, strings.TrimSpace(req.DefaultWechat), strings.TrimSpace(req.DefaultPhone)))
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, file *multipart.FileHeader) error {
	if file == nil {
		return ErrAvatarFileRequired
	}
	if s.uploader == nil {
		return ErrStorageNotConfigured
	}
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return err
	}

	url, err := s.uploader.UploadAvatar(ctx, userID, file)
	if err != nil {
		return err
	}
	return normalizeUserRepoError(s.userRepo.UpdateAvatarURL(ctx, userID, url))
}

func (s *UserService) ensureUserExists(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return normalizeUserRepoError(err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	return nil
}

func normalizeUserRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, mysql.ErrUserNotFound) || errors.Is(err, ErrUserNotFound) {
		return ErrUserNotFound
	}
	return err
}
