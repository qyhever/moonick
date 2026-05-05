package storage

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"moonick/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type putObjectStub struct {
	called        bool
	bucket        string
	key           string
	contentType   string
	contentLength int64
}

func (s *putObjectStub) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	s.called = true
	s.bucket = aws.ToString(params.Bucket)
	s.key = aws.ToString(params.Key)
	s.contentType = aws.ToString(params.ContentType)
	if params.ContentLength != nil {
		s.contentLength = *params.ContentLength
	}
	return &s3.PutObjectOutput{}, nil
}

func TestR2UploadUploadsObjectAndReturnsPublicURL(t *testing.T) {
	file := makeMultipartFileHeader(t, "avatar.png", "image/png", []byte("avatar"))
	client := &putObjectStub{}
	storage := &R2{
		client:        client,
		bucketName:    "moonick",
		publicBaseURL: "https://cdn.example.com/moonick",
	}

	url, err := storage.Upload(context.Background(), "avatars/1001/test.png", file)
	if err != nil {
		t.Fatalf("upload returned error: %v", err)
	}
	if !client.called {
		t.Fatal("expected PutObject to be called")
	}
	if client.bucket != "moonick" {
		t.Fatalf("expected bucket moonick, got %q", client.bucket)
	}
	if client.key != "avatars/1001/test.png" {
		t.Fatalf("expected key avatars/1001/test.png, got %q", client.key)
	}
	if client.contentType != "image/png" {
		t.Fatalf("expected content type image/png, got %q", client.contentType)
	}
	if client.contentLength != int64(len("avatar")) {
		t.Fatalf("expected content length %d, got %d", len("avatar"), client.contentLength)
	}
	if url != "https://cdn.example.com/moonick/avatars/1001/test.png" {
		t.Fatalf("unexpected url %q", url)
	}
}

func makeMultipartFileHeader(t *testing.T, filename, contentType string, body []byte) *multipart.FileHeader {
	t.Helper()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/", payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(len(body)) + 1024); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	t.Cleanup(func() {
		if req.MultipartForm != nil {
			_ = req.MultipartForm.RemoveAll()
		}
	})

	file, header, err := req.FormFile("file")
	if err != nil {
		t.Fatalf("read form file: %v", err)
	}
	_ = file.Close()
	header.Header.Set("Content-Type", contentType)
	return header
}

func TestBuildBaseURLUsesR2EndpointWhenConfigured(t *testing.T) {
	url := buildBaseURL(config.R2Config{
		AccountID:  "account",
		BucketName: "bucket",
	})
	if !strings.Contains(url, "account.r2.cloudflarestorage.com/bucket") {
		t.Fatalf("unexpected base url %q", url)
	}
}

func TestBuildBaseURLPrefersConfiguredPublicBaseURL(t *testing.T) {
	url := buildBaseURL(config.R2Config{
		AccountID:     "account",
		BucketName:    "bucket",
		PublicBaseURL: "https://pub-c986c455c9884a098d2751147be48ba8.r2.dev/",
	})
	if url != "https://pub-c986c455c9884a098d2751147be48ba8.r2.dev" {
		t.Fatalf("unexpected public base url %q", url)
	}
}
