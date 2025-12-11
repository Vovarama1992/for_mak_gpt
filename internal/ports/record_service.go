package ports

import "context"

type RecordService interface {
	AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error)
	AddImage(ctx context.Context, botID string, telegramID int64, role, imageURL string) (int64, error)

	GetHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
	GetFittingHistory(ctx context.Context, botID string, telegramID int64) ([]Record, error)
	RecalcHistoryState(ctx context.Context, botID string, telegramID int64) error

	ListUsers(ctx context.Context) ([]UserBots, error)
	DeleteAll(ctx context.Context) error

	// НОВОЕ — очистка истории конкретного юзера в конкретном боте
	DeleteUserHistory(ctx context.Context, botID string, telegramID int64) error
}
