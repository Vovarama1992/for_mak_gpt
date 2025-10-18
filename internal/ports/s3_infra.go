package ports

import (
	"context"
	"io"
)

// Низкоуровневый клиент к S3
type S3Client interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) (publicURL string, err error)
}
