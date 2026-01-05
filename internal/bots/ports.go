package bots

import (
	"context"
	"io"
)

type Repo interface {
	ListAll(ctx context.Context) ([]*BotConfig, error)
	Get(ctx context.Context, botID string) (*BotConfig, error)
	Update(ctx context.Context, cfg *UpdateInput) (*BotConfig, error)
}

type Service interface {
	ListAll(ctx context.Context) ([]*BotConfig, error)
	Get(ctx context.Context, botID string) (*BotConfig, error)
	Update(ctx context.Context, cfg *UpdateInput) (*BotConfig, error)

	// загрузка приветственного видео → S3 → запись в bot_configs
	UploadWelcomeVideo(ctx context.Context, botID string, file io.Reader, filename string) (string, error)
}

type BotConfig struct {
	BotID            string `json:"bot_id"`
	Token            string `json:"token"`
	Model            string `json:"model"`
	TextStylePrompt  string `json:"text_style_prompt"`
	VoiceStylePrompt string `json:"voice_style_prompt"`
	PhotoStylePrompt string `json:"photo_style_prompt"`
	VoiceID          string `json:"voice_id"`

	WelcomeText        *string `json:"welcome_text"`
	TariffText         *string `json:"tariff_text"`
	AfterContinueText  *string `json:"after_continue_text"`
	NoVoiceMinutesText *string `json:"no_voice_minutes_text"`
	WelcomeVideo       *string `json:"welcome_video_url"`
}

type UpdateInput struct {
	BotID              string
	Model              *string
	TextStylePrompt    *string
	VoiceStylePrompt   *string
	PhotoStylePrompt   *string
	VoiceID            *string
	WelcomeText        *string
	TariffText         *string
	AfterContinueText  *string
	NoVoiceMinutesText *string

	// INTERNAL USE ONLY
	WelcomeVideo *string
}
