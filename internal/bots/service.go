package bots

import "context"

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) ListAll(ctx context.Context) ([]*BotConfig, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) Get(ctx context.Context, botID string) (*BotConfig, error) {
	return s.repo.Get(ctx, botID)
}

func (s *service) Update(ctx context.Context, in *UpdateInput) (*BotConfig, error) {
	return s.repo.Update(ctx, in)
}
