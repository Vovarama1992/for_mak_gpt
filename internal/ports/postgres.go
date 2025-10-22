package ports

import (
	"context"
	"time"
)

// DTO для истории
type Record struct {
	ID         int64
	TelegramID int64
	BotID      *string // новое поле
	UserRef    *int64
	Role       string
	Type       string
	Text       *string
	ImageURL   *string
	CreatedAt  time.Time
}

// Репозиторий Postgres
type RecordRepo interface {
	CreateText(ctx context.Context, botID *string, telegramID int64, role, text string) (int64, error)
	CreateImage(ctx context.Context, botID *string, telegramID int64, role, imageURL string) (int64, error)
	GetHistory(ctx context.Context, botID *string, telegramID int64) ([]Record, error)
}
