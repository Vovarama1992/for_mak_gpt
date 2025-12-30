package classes

import "context"

type service struct {
	repo ClassRepo
}

func NewClassService(repo ClassRepo) ClassService {
	return &service{repo: repo}
}

//
// classes
//

func (s *service) CreateClass(ctx context.Context, botID string, grade string) (*Class, error) {
	return s.repo.CreateClass(ctx, botID, grade)
}

func (s *service) ListClasses(ctx context.Context, botID string) ([]*Class, error) {
	return s.repo.ListClasses(ctx, botID)
}

func (s *service) GetClassByID(ctx context.Context, botID string, id int) (*Class, error) {
	return s.repo.GetClassByID(ctx, botID, id)
}

func (s *service) UpdateClass(ctx context.Context, botID string, id int, grade string) error {
	return s.repo.UpdateClass(ctx, botID, id, grade)
}

func (s *service) DeleteClass(ctx context.Context, botID string, id int) error {
	return s.repo.DeleteClass(ctx, botID, id)
}

//
// prompts
//

func (s *service) CreatePrompt(ctx context.Context, botID string, classID int, prompt string) (*ClassPrompt, error) {
	return s.repo.CreatePrompt(ctx, botID, classID, prompt)
}

func (s *service) UpdatePrompt(ctx context.Context, botID string, id int, prompt string) error {
	return s.repo.UpdatePrompt(ctx, botID, id, prompt)
}

func (s *service) DeletePrompt(ctx context.Context, botID string, id int) error {
	return s.repo.DeletePrompt(ctx, botID, id)
}

func (s *service) GetPromptByClassID(ctx context.Context, botID string, classID int) (*ClassPrompt, error) {
	return s.repo.GetPromptByClassID(ctx, botID, classID)
}

//
// user → class
//

func (s *service) SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error {
	return s.repo.SetUserClass(ctx, botID, telegramID, classID)
}

func (s *service) GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error) {
	return s.repo.GetUserClass(ctx, botID, telegramID)
}

func (s *service) DeleteUserClass(ctx context.Context, botID string, telegramID int64) error {
	return s.repo.DeleteUserClass(ctx, botID, telegramID)
}
