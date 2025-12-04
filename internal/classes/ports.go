package classes

import "context"

// --- базовая сущность класса ---
type Class struct {
	ID     int          `json:"id"`
	Grade  string       `json:"grade"`
	Prompt *ClassPrompt `json:"prompt,omitempty"`
}

// --- промпт для класса ---
type ClassPrompt struct {
	ID      int    `json:"id"`
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
	CreateClass(ctx context.Context, grade string) (*Class, error)
	ListClasses(ctx context.Context) ([]*Class, error)
	GetClassByID(ctx context.Context, id int) (*Class, error)

	// class_prompts
	CreatePrompt(ctx context.Context, classID int, prompt string) (*ClassPrompt, error)
	UpdatePrompt(ctx context.Context, id int, prompt string) error
	DeletePrompt(ctx context.Context, id int) error
	GetPromptByClassID(ctx context.Context, classID int) (*ClassPrompt, error)

	// user_classes
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}

type ClassService interface {
	// classes
	CreateClass(ctx context.Context, grade string) (*Class, error)
	ListClasses(ctx context.Context) ([]*Class, error)
	GetClassByID(ctx context.Context, id int) (*Class, error)

	// prompts
	CreatePrompt(ctx context.Context, classID int, prompt string) (*ClassPrompt, error)
	UpdatePrompt(ctx context.Context, id int, prompt string) error
	DeletePrompt(ctx context.Context, id int) error
	GetPromptByClassID(ctx context.Context, classID int) (*ClassPrompt, error)

	// user → class
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}
