package prompts

import "context"

type Repo interface {
	ListAll(ctx context.Context) ([]*Prompt, error)
	Update(ctx context.Context, botID, prompt string) (*Prompt, error)
	GetByBotID(ctx context.Context, botID string) (string, error) // ← добавляем
}

type Service interface {
	ListAll(ctx context.Context) ([]*Prompt, error)
	Update(ctx context.Context, botID, prompt string) (*Prompt, error)
	GetByBotID(ctx context.Context, botID string) (string, error)
}

type Prompt struct {
	BotID  string `json:"bot_id"`
	Prompt string `json:"prompt"`
}
