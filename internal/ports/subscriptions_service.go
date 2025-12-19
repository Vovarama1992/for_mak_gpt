package ports

import (
	"context"
	"time"
)

type SubscriptionService interface {
	// создание подписки (создаёт запись и создаёт платёж в Юкассе)
	Create(ctx context.Context, botID string, telegramID int64, planCode string) (paymentURL string, err error)

	// активация по вебхуку (по payment_id)
	Activate(ctx context.Context, paymentID string) error

	ActivateTrial(ctx context.Context, botID string, telegramID int64, planCode string) error

	// получение текущего статуса подписки пользователя
	GetStatus(ctx context.Context, botID string, telegramID int64) (string, error)

	// получение подписки целиком
	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)

	// начисление минут по пакету
	AddMinutesFromPackage(
		ctx context.Context,
		botID string,
		telegramID int64,
		packageID int64,
	) error

	// списание голосовых минут. ok=false — если не хватило
	UseVoiceMinutes(ctx context.Context, botID string, telegramID int64, used float64) (ok bool, err error)

	// список всех подписок (например, для админки)
	ListAll(ctx context.Context) ([]*Subscription, error)

	// очистка всех pending старше olderThan
	CleanupPending(ctx context.Context, olderThan time.Duration) error
	Delete(ctx context.Context, botID string, telegramID int64) error
	CleanupExpiredTrials(ctx context.Context, botID string) error
}
