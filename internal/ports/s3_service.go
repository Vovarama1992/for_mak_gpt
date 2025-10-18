package ports

import (
	"context"
	"mime/multipart"
)

// Высокоуровневый S3-сервис (правильные ключи/пути)
type S3Service interface {
	// сохраняет файл и возвращает публичный URL
	SaveImage(ctx context.Context, telegramID int64, file multipart.File, filename string, contentType string) (publicURL string, err error)
	// формирует ключ вида telegram_id/yyyy-mm-dd/filename
	ObjectKey(telegramID int64, filename string) string
}
