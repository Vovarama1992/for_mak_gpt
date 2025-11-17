package ports

import (
	"context"
	"io"
)

// Бизнес-сервис для работы с записями
type RecordService interface {
	// уже есть
	AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error)
	// теперь возвращаем id и publicURL
	AddImage(ctx context.Context, botID string, telegramID int64, role string, file io.Reader, filename, contentType string) (int64, string, error)
	GetHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
	ListUsers(ctx context.Context) ([]UserBots, error)

	// новые — после каждого AddText/AddImage
	RecalcHistoryState(ctx context.Context, botID string, telegramID int64) error
	GetFittingHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
}
