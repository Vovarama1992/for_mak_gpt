package user

import "context"

// Infra — работа с БД
type Infra interface {
	ResetUserSettings(ctx context.Context, botID string, telegramID int64) error
}

// Service — бизнес-операции
type Service interface {
	ResetUserSettings(ctx context.Context, botID string, telegramID int64) error
}
