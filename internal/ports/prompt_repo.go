package ports

import "context"

type PromptRepo interface {
	GetByBotID(ctx context.Context, botID string) (string, error)
}