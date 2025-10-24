package ports

import "context"

type SubscriptionService interface {
	Create(ctx context.Context, botID string, telegramID int64, planCode string) error
	Activate(ctx context.Context, botID string, telegramID int64) error
	GetStatus(ctx context.Context, botID string, telegramID int64) (string, error)
	ListAll(ctx context.Context) ([]*Subscription, error)
}
