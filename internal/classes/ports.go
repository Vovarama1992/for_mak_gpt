package classes

import "context"

// --- базовая сущность класса ---
type Class struct {
	ID    int    `json:"id"`
	BotID string `json:"bot_id"`
	Grade string `json:"grade"`

	Prompt *ClassPrompt `json:"prompt,omitempty"`
}

// --- промпт для класса ---
type ClassPrompt struct {
	ID      int    `json:"id"`
	BotID   string `json:"bot_id"`
	ClassID int    `json:"class_id"` // FK → classes.id
	Prompt  string `json:"prompt"`
}

// --- связка пользователя с классом ---
type UserClass struct {
	BotID      string `json:"bot_id"`
	TelegramID int64  `json:"telegram_id"`
	ClassID    int    `json:"class_id"`
}

// --- интерфейс репозитория ---
type ClassRepo interface {
	// classes
	CreateClass(ctx context.Context, botID string, grade string) (*Class, error)
	ListClasses(ctx context.Context) ([]*Class, error)
	GetClassByID(ctx context.Context, botID string, id int) (*Class, error)
	UpdateClass(ctx context.Context, botID string, id int, grade string) error
	DeleteClass(ctx context.Context, botID string, id int) error

	// class_prompts
	CreatePrompt(ctx context.Context, botID string, classID int, prompt string) (*ClassPrompt, error)
	UpdatePrompt(ctx context.Context, botID string, id int, prompt string) error
	DeletePrompt(ctx context.Context, botID string, id int) error
	GetPromptByClassID(ctx context.Context, botID string, classID int) (*ClassPrompt, error)

	// user_classes
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}

type ClassService interface {
	// classes
	CreateClass(ctx context.Context, botID string, grade string) (*Class, error)
	ListClasses(ctx context.Context) ([]*Class, error)
	GetClassByID(ctx context.Context, botID string, id int) (*Class, error)
	UpdateClass(ctx context.Context, botID string, id int, grade string) error
	DeleteClass(ctx context.Context, botID string, id int) error

	// prompts
	CreatePrompt(ctx context.Context, botID string, classID int, prompt string) (*ClassPrompt, error)
	UpdatePrompt(ctx context.Context, botID string, id int, prompt string) error
	DeletePrompt(ctx context.Context, botID string, id int) error
	GetPromptByClassID(ctx context.Context, botID string, classID int) (*ClassPrompt, error)

	// user → class
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}
