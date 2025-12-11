package ports

import (
	"context"
	"time"
)

// DTO для истории сообщений
type Record struct {
	ID         int64
	TelegramID int64
	BotID      string
	UserRef    *int64
	Role       string
	Type       string
	Text       *string
	ImageURL   *string
	CreatedAt  time.Time
}

// DTO для списка пользователей и их ботов
type UserBots struct {
	TelegramID int64    `json:"telegram_id"`
	Bots       []string `json:"bots"`
}

type HistoryState struct {
	BotID        string
	TelegramID   int64
	LastNRecords int // сколько последних записей вмещается
	TotalTokens  int // суммарный вес их в токенах
	UpdatedAt    time.Time
}

// Репозиторий Postgres
type RecordRepo interface {
	// уже есть
	CreateText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error)
	CreateImage(ctx context.Context, botID string, telegramID int64, role, imageURL string) (int64, error)
	GetHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
	ListUsers(ctx context.Context) ([]UserBots, error)

	// новые
	UpsertHistoryState(ctx context.Context, botID string, telegramID int64, lastN, totalTokens int) error
	GetHistoryState(ctx context.Context, botID string, telegramID int64) (lastN, totalTokens int, err error)
	GetLastNRecords(ctx context.Context, botID string, telegramID int64, n int) ([]Record, error)
	DeleteAll(ctx context.Context) error

	DeleteByUser(ctx context.Context, botID string, telegramID int64) error
}
