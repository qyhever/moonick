package service

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"moonick/internal/model/entity"
	"moonick/internal/model/request"
)

func TestUserService_UpdateAvatarRollbackOnUploadError(t *testing.T) {
	repo := &userRepoStub{
		user: &entity.User{ID: 1001, Email: "user@example.com", Phone: "13800138000", Nickname: "测试用户", Status: "active"},
	}
	svc := newUserServiceForTest(repo, &uploadStub{err: errors.New("r2 down")})

	err := svc.UpdateAvatar(context.Background(), 1001, fakeFileHeader(t, "avatar.png"))
	if err == nil {
		t.Fatal("expected update avatar error")
	}
	if err.Error() != "r2 down" {
		t.Fatalf("expected r2 down error, got %v", err)
	}
	if repo.avatarUpdated {
		t.Fatal("expected avatar update to rollback on upload error")
	}
}

func TestUserService_UpdateAvatarReturnsUserNotFoundBeforeUpload(t *testing.T) {
	repo := &userRepoStub{}
	uploader := &uploadStub{}
	svc := newUserServiceForTest(repo, uploader)

	err := svc.UpdateAvatar(context.Background(), 9999, fakeFileHeader(t, "avatar.png"))
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if uploader.called {
		t.Fatal("expected uploader not to be called when user does not exist")
	}
}

func TestUserService_UpdateProfileReturnsUserNotFound(t *testing.T) {
	svc := newUserServiceForTest(&userRepoStub{}, &uploadStub{})

	err := svc.UpdateProfile(context.Background(), 9999, request.UpdateProfileRequest{Nickname: "新昵称"})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateContactReturnsUserNotFound(t *testing.T) {
	svc := newUserServiceForTest(&userRepoStub{}, &uploadStub{})

	err := svc.UpdateContact(context.Background(), 9999, request.UpdateContactRequest{DefaultPhone: "13800138000"})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

type uploadStub struct {
	err    error
	called bool
}

func (s *uploadStub) UploadAvatar(_ context.Context, _ int64, _ *multipart.FileHeader) (string, error) {
	s.called = true
	if s.err != nil {
		return "", s.err
	}
	return "https://cdn.example/avatar.png", nil
}

type userRepoStub struct {
	user          *entity.User
	avatarUpdated bool
}

func (s *userRepoStub) FindByID(_ context.Context, id int64) (*entity.User, error) {
	if s.user == nil || s.user.ID != id {
		return nil, nil
	}
	copied := *s.user
	return &copied, nil
}

func (s *userRepoStub) UpdateProfile(_ context.Context, userID int64, nickname string) error {
	if s.user == nil || s.user.ID != userID {
		return ErrUserNotFound
	}
	s.user.Nickname = nickname
	return nil
}

func (s *userRepoStub) UpdateContact(_ context.Context, userID int64, defaultWechat, defaultPhone string) error {
	if s.user == nil || s.user.ID != userID {
		return ErrUserNotFound
	}
	s.user.DefaultWechat = defaultWechat
	s.user.DefaultPhone = defaultPhone
	return nil
}

func (s *userRepoStub) UpdateAvatarURL(_ context.Context, userID int64, avatarURL string) error {
	if s.user == nil || s.user.ID != userID {
		return ErrUserNotFound
	}
	s.user.AvatarURL = avatarURL
	s.avatarUpdated = true
	return nil
}

func newUserServiceForTest(repo *userRepoStub, uploader avatarUploader) *UserService {
	return NewUserService(repo, uploader)
}

func fakeFileHeader(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte("avatar"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len("avatar")),
	}
}
