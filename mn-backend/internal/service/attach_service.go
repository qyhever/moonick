package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"moonick/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type AttachService struct {
	client *s3.Client
}

func NewAttachService() (*AttachService, error) {
	globalConfig := config.GetConfig()
	if globalConfig == nil {
		return nil, fmt.Errorf("global config is not initialized")
	}

	r2Config := globalConfig.R2

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r2Config.AccessKeyID,
			r2Config.AccessKeySecret,
			"",
		)),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2Config.AccountID))
	})

	return &AttachService{
		client: client,
	}, nil
}

// UploadFile 上传文件到 R2
func (s *AttachService) UploadFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 按日期分目录，避免单目录文件过多
	objectKey := fmt.Sprintf("%s/%s", time.Now().Format("2006-01-02"), newFileName)

	bucketName := config.GetConfig().R2.BucketName
	if bucketName == "" {
		return "", fmt.Errorf("bucket name is not configured")
	}

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          src,
		ContentLength: aws.Int64(file.Size),
		ContentType:   aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to R2: %v", err)
	}

	return objectKey, nil
}

// DeleteFile 删除 R2 中的文件
func (s *AttachService) DeleteFile(key string) error {
	bucketName := config.GetConfig().R2.BucketName
	if bucketName == "" {
		return fmt.Errorf("bucket name is not configured")
	}

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %v", err)
	}
	return nil
}

// ListFiles 列出 R2 中的文件
func (s *AttachService) ListFiles() (string, []string, error) {
	bucketName := config.GetConfig().R2.BucketName
	if bucketName == "" {
		return bucketName, nil, fmt.Errorf("bucket name is not configured")
	}

	output, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return bucketName, nil, fmt.Errorf("failed to list files: %v", err)
	}

	var files []string
	for _, object := range output.Contents {
		files = append(files, *object.Key)
	}
	return bucketName, files, nil
}
