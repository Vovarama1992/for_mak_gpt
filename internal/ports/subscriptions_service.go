package ports

import "context"

type SubscriptionService interface {
	// создание подписки (создаёт запись и создаёт платёж в Юкассе)
	Create(ctx context.Context, botID string, telegramID int64, planCode string) (paymentURL string, err error)

	// активация по вебхуку (по payment_id)
	Activate(ctx context.Context, paymentID string) error

	// получение текущего статуса подписки пользователя
	GetStatus(ctx context.Context, botID string, telegramID int64) (string, error)

	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)

	AddMinutesFromPackage(
		ctx context.Context,
		botID string,
		telegramID int64,
		packageID int64,
	) error

	// списание голосовых минут. ok=false, если не хватило
	UseVoiceMinutes(ctx context.Context, botID string, telegramID int64, used float64) (ok bool, err error)

	// список всех подписок (для админки)
	ListAll(ctx context.Context) ([]*Subscription, error)
}
