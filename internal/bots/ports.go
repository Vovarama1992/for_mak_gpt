package bots

import "context"

type Repo interface {
	ListAll(ctx context.Context) ([]*BotConfig, error)
	Get(ctx context.Context, botID string) (*BotConfig, error)
	Update(ctx context.Context, cfg *UpdateInput) (*BotConfig, error)
}

type Service interface {
	ListAll(ctx context.Context) ([]*BotConfig, error)
	Get(ctx context.Context, botID string) (*BotConfig, error)
	Update(ctx context.Context, cfg *UpdateInput) (*BotConfig, error)
}

type BotConfig struct {
	BotID       string `json:"bot_id"`
	Token       string `json:"token"`
	Model       string `json:"model"`
	StylePrompt string `json:"style_prompt"`
	VoiceID     string `json:"voice_id"`
}

type UpdateInput struct {
	BotID       string
	Model       *string
	StylePrompt *string
	VoiceID     *string
}
