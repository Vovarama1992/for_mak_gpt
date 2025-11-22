package error_notificator

import "context"

type Service struct {
	infra Notificator
}

func NewService(infra Notificator) *Service {
	return &Service{infra: infra}
}

func (s *Service) Notify(ctx context.Context, botID string, err error, details string) error {
	return s.infra.Notify(ctx, botID, err, details)
}
