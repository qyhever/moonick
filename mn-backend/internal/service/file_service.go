package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type objectStorage interface {
	Upload(ctx context.Context, key string, file *multipart.FileHeader) (string, error)
}

type FileService struct {
	storage objectStorage
}

func NewFileService(storage objectStorage) *FileService {
	return &FileService{storage: storage}
}

func (s *FileService) UploadAvatar(ctx context.Context, userID int64, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", ErrAvatarFileRequired
	}
	if s.storage == nil {
		return "", ErrStorageNotConfigured
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".bin"
	}

	key := fmt.Sprintf("avatars/%d/%d%s", userID, time.Now().UnixNano(), ext)
	return s.storage.Upload(ctx, key, file)
}
