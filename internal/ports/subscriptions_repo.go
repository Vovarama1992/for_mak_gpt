package ports

import (
	"context"
	"time"
)

type Subscription struct {
	ID                int
	BotID             string
	TelegramID        int64
	PlanID            int
	Status            string
	StartedAt         time.Time
	ExpiresAt         *time.Time
	UpdatedAt         time.Time
	YookassaPaymentID string
}

type SubscriptionRepo interface {
	// создание новой подписки
	Create(ctx context.Context, s *Subscription) error

	// получение подписки по PaymentID (для вебхука)
	GetByPaymentID(ctx context.Context, paymentID string) (*Subscription, error)

	// обновление статуса по ID
	UpdateStatus(ctx context.Context, id int, status string) error

	// получение подписки по боту и телеграму (для API и админки)
	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)

	// список всех подписок (для админки)
	ListAll(ctx context.Context) ([]*Subscription, error)
}
