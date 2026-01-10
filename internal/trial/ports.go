package trial

import "context"

type RepoInf interface {
	Exists(ctx context.Context, botID string, telegramID int64) (bool, error)
	Create(ctx context.Context, botID string, telegramID int64) error
	Delete(
		ctx context.Context,
		botID string,
		telegramID int64,
	) error
}
