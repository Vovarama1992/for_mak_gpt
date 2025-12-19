package notificator

import "context"

type Notificator interface {
	Notify(ctx context.Context, botID string, err error, details string) error
	UserNotify(ctx context.Context, botID string, chatID int64, text string) error
}
