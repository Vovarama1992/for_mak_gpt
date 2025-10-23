package ports

import (
	"context"
	"io"
)

// Низкоуровневый клиент к S3
type S3Client interface {
	PutObject(context.Context, string, string, io.Reader, int64, string) (string, error)
}
