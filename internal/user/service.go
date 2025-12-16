package user

import "context"

type service struct {
	infra Infra
}

func NewService(infra Infra) Service {
	return &service{infra: infra}
}

func (s *service) ResetUserSettings(
	ctx context.Context,
	botID string,
	telegramID int64,
) error {
	return s.infra.ResetUserSettings(ctx, botID, telegramID)
}
