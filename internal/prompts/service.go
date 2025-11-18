package prompts

import "context"

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) ListAll(ctx context.Context) ([]*Prompt, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) Update(ctx context.Context, botID, prompt string) (*Prompt, error) {
	return s.repo.Update(ctx, botID, prompt)
}

func (s *service) GetByBotID(ctx context.Context, botID string) (string, error) {
	return s.repo.GetByBotID(ctx, botID)
}
