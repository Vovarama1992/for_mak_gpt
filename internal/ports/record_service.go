package ports

import (
	"context"
	"mime/multipart"
)

// Бизнес-сервис для работы с записями
type RecordService interface {
	AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error)
	AddImage(ctx context.Context, botID string, telegramID int64, role string, file multipart.File, filename string, contentType string) (int64, error)
	GetHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
}
