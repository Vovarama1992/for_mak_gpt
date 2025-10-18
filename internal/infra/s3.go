package infra

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Client struct {
	client *minio.Client
	bucket string
	host   string
}

func NewS3Client() (ports.S3Client, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init S3 client: %w", err)
	}

	// проверим, что бакет существует
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket %q does not exist", bucket)
	}

	return &s3Client{
		client: client,
		bucket: bucket,
		host:   fmt.Sprintf("https://%s", endpoint),
	}, nil
}

// PutObject загружает файл и возвращает публичный URL
func (s *s3Client) PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: map[string]string{"uploaded-at": time.Now().Format(time.RFC3339)},
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	publicURL := s.buildPublicURL(key)
	return publicURL, nil
}

func (s *s3Client) buildPublicURL(key string) string {
	escapedKey := url.PathEscape(filepath.ToSlash(key))
	return fmt.Sprintf("%s/%s/%s", s.host, s.bucket, escapedKey)
}
