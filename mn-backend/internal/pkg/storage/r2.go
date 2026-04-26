package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"moonick/internal/config"
)

type R2 struct {
	publicBaseURL string
}

func NewR2(cfg config.R2Config) *R2 {
	return &R2{
		publicBaseURL: buildBaseURL(cfg),
	}
}

func (r *R2) Upload(_ context.Context, key string, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file header is nil")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if _, err := io.Copy(io.Discard, src); err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(r.publicBaseURL, "/")
	return baseURL + "/" + strings.TrimLeft(key, "/"), nil
}

func buildBaseURL(cfg config.R2Config) string {
	if cfg.AccountID != "" && cfg.BucketName != "" {
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", cfg.AccountID, cfg.BucketName)
	}
	if cfg.BucketName != "" {
		return fmt.Sprintf("https://r2.invalid/%s", cfg.BucketName)
	}
	return "https://r2.invalid/moonick"
}
