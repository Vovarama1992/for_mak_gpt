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

func (s *service) CreateClass(ctx context.Context, grade string) (*Class, error) {
	return s.repo.CreateClass(ctx, grade)
}

func (s *service) ListClasses(ctx context.Context) ([]*Class, error) {
	return s.repo.ListClasses(ctx)
}

func (s *service) GetClassByID(ctx context.Context, id int) (*Class, error) {
	return s.repo.GetClassByID(ctx, id)
}

func (s *service) UpdateClass(ctx context.Context, id int, grade string) error {
	return s.repo.UpdateClass(ctx, id, grade)
}

func (s *service) DeleteClass(ctx context.Context, id int) error {
	return s.repo.DeleteClass(ctx, id)
}

//
// prompts
//

func (s *service) CreatePrompt(ctx context.Context, classID int, prompt string) (*ClassPrompt, error) {
	return s.repo.CreatePrompt(ctx, classID, prompt)
}

func (s *service) UpdatePrompt(ctx context.Context, id int, prompt string) error {
	return s.repo.UpdatePrompt(ctx, id, prompt)
}

func (s *service) DeletePrompt(ctx context.Context, id int) error {
	return s.repo.DeletePrompt(ctx, id)
}

func (s *service) GetPromptByClassID(ctx context.Context, classID int) (*ClassPrompt, error) {
	return s.repo.GetPromptByClassID(ctx, classID)
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
