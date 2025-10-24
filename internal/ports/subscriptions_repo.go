package ports

import (
	"context"
	"time"
)

type Subscription struct {
	ID         int
	BotID      string
	TelegramID int64
	PlanID     int
	Status     string
	StartedAt  time.Time
	ExpiresAt  *time.Time
	UpdatedAt  time.Time
}

type SubscriptionRepo interface {
	Create(ctx context.Context, s *Subscription) error
	UpdateStatus(ctx context.Context, botID string, telegramID int64, status string) error
	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)
	ListAll(ctx context.Context) ([]*Subscription, error) // для админки
}
