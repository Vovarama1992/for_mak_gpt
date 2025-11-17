package ports

import (
	"context"
	"mime/multipart"
)

// Бизнес-сервис для работы с записями
type RecordService interface {
	// уже есть
	AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error)
	AddImage(ctx context.Context, botID string, telegramID int64, role string, file multipart.File, filename string, contentType string) (int64, error)
	GetHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
	ListUsers(ctx context.Context) ([]UserBots, error)

	// новые — после каждого AddText/AddImage
	RecalcHistoryState(ctx context.Context, botID string, telegramID int64) error
	GetFittingHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
}
