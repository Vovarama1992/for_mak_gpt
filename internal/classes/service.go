package classes

import "context"

type ClassService interface {
	// prompts
	CreatePrompt(ctx context.Context, p *ClassPrompt) error
	UpdatePrompt(ctx context.Context, p *ClassPrompt) error
	DeletePrompt(ctx context.Context, id int) error
	GetPromptByID(ctx context.Context, id int) (*ClassPrompt, error)
	GetPromptByClass(ctx context.Context, class int) (*ClassPrompt, error)
	ListPrompts(ctx context.Context) ([]*ClassPrompt, error)

	// user → class
	SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error
	GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error)
	DeleteUserClass(ctx context.Context, botID string, telegramID int64) error
}

type service struct {
	repo ClassRepo
}

func NewClassService(repo ClassRepo) ClassService {
	return &service{repo: repo}
}

// prompts

func (s *service) CreatePrompt(ctx context.Context, p *ClassPrompt) error {
	return s.repo.CreatePrompt(ctx, p)
}
func (s *service) UpdatePrompt(ctx context.Context, p *ClassPrompt) error {
	return s.repo.UpdatePrompt(ctx, p)
}
func (s *service) DeletePrompt(ctx context.Context, id int) error {
	return s.repo.DeletePrompt(ctx, id)
}
func (s *service) GetPromptByID(ctx context.Context, id int) (*ClassPrompt, error) {
	return s.repo.GetPromptByID(ctx, id)
}
func (s *service) GetPromptByClass(ctx context.Context, class int) (*ClassPrompt, error) {
	return s.repo.GetPromptByClass(ctx, class)
}
func (s *service) ListPrompts(ctx context.Context) ([]*ClassPrompt, error) {
	return s.repo.ListPrompts(ctx)
}

// user → class

func (s *service) SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error {
	return s.repo.SetUserClass(ctx, botID, telegramID, classID)
}
func (s *service) GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error) {
	return s.repo.GetUserClass(ctx, botID, telegramID)
}
func (s *service) DeleteUserClass(ctx context.Context, botID string, telegramID int64) error {
	return s.repo.DeleteUserClass(ctx, botID, telegramID)
}
