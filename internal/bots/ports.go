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
	BotID            string `json:"bot_id"`
	Token            string `json:"token"`
	Model            string `json:"model"`
	TextStylePrompt  string `json:"text_style_prompt"`
	VoiceStylePrompt string `json:"voice_style_prompt"`
	VoiceID          string `json:"voice_id"`
}

type UpdateInput struct {
	BotID            string
	Model            *string
	TextStylePrompt  *string
	VoiceStylePrompt *string
	VoiceID          *string
}
