package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"moonick/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Uploader interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type R2 struct {
	client        s3Uploader
	bucketName    string
	publicBaseURL string
}

func NewR2(cfg config.R2Config) *R2 {
	r2 := &R2{
		bucketName:    strings.TrimSpace(cfg.BucketName),
		publicBaseURL: buildBaseURL(cfg),
	}

	if strings.TrimSpace(cfg.AccountID) == "" || strings.TrimSpace(cfg.BucketName) == "" {
		return r2
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.AccessKeySecret,
			"",
		)),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return r2
	}

	r2.client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})

	return r2
}

func (r *R2) Upload(ctx context.Context, key string, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file header is nil")
	}
	if strings.TrimSpace(r.bucketName) == "" {
		return "", fmt.Errorf("bucket name is not configured")
	}
	if r.client == nil {
		return "", fmt.Errorf("r2 client is not initialized")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(key),
		Body:          src,
		ContentLength: aws.Int64(file.Size),
		ContentType:   aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(r.publicBaseURL, "/")
	return baseURL + "/" + strings.TrimLeft(key, "/"), nil
}

func buildBaseURL(cfg config.R2Config) string {
	if strings.TrimSpace(cfg.PublicBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	}
	if cfg.AccountID != "" && cfg.BucketName != "" {
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", cfg.AccountID, cfg.BucketName)
	}
	if cfg.BucketName != "" {
		return fmt.Sprintf("https://r2.invalid/%s", cfg.BucketName)
	}
	return "https://r2.invalid/moonick"
}
