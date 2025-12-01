package classes

import "context"

type ClassPrompt struct {
	ID     int    `json:"id"`
	Class  int    `json:"class"` // 1..11
	Prompt string `json:"prompt"`
}

type UserClass struct {
	BotID      string `json:"bot_id"`
	TelegramID int64  `json:"telegram_id"`
	ClassID    int    `json:"class_id"`
}

type ClassRepo interface {
	// class_prompts
	CreatePrompt(ctx context.Context, p *ClassPrompt) error
	UpdatePrompt(ctx context.Context, p *ClassPrompt) error
	DeletePrompt(ctx context.Context, id int) error
	GetPromptByID(ctx context.Context, id int) (*ClassPrompt, error)
	GetPromptByClass(ctx context.Context, class int) (*ClassPrompt, error)
	ListPrompts(ctx context.Context) ([]*ClassPrompt, error)

	// user_classes
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}
