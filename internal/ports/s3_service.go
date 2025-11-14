package ports

import (
	"context"
	"io"
)

type S3Service interface {
	SaveImage(ctx context.Context, botID string, telegramID int64, file io.Reader, filename, contentType string) (string, error)
}
