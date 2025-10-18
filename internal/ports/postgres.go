package ports

import (
	"context"
	"time"
)

// DTO для истории (достаточно для слоя выше)
type Record struct {
	ID         int64
	TelegramID int64
	UserRef    *int64
	Role       string  // "user" | "tutor"
	Type       string  // "text" | "image"
	Text       *string // nullable
	ImageURL   *string // nullable
	CreatedAt  time.Time
}

// Репозиторий Postgres
type RecordRepo interface {
	CreateText(ctx context.Context, telegramID int64, role, text string) (int64, error)
	CreateImage(ctx context.Context, telegramID int64, role, imageURL string) (int64, error)
	GetHistory(ctx context.Context, telegramID int64) ([]Record, error)
}
