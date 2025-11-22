package error_notificator

import "context"

type Notificator interface {
	// Notify — отправляет сообщение об ошибке админу
	Notify(ctx context.Context, botID string, err error, details string) error
}
