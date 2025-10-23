package ports

import (
	"context"
	"mime/multipart"
)

type S3Service interface {
	ObjectKey(telegramID int64, filename string) string
	SaveImage(ctx context.Context, botID string, telegramID int64, file multipart.File, filename, contentType string) (string, error)
}
