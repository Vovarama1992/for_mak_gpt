package domain

import (
	"context"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type SubscriptionService struct {
	repo ports.SubscriptionRepo
}

func NewSubscriptionService(repo ports.SubscriptionRepo) ports.SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, botID string, telegramID int64, planCode string) error {
	// на этом этапе просто создаём запись с pending
	sub := &ports.Subscription{
		BotID:      botID,
		TelegramID: telegramID,
		Status:     "pending",
		StartedAt:  time.Now(),
	}
	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) Activate(ctx context.Context, botID string, telegramID int64) error {
	return s.repo.UpdateStatus(ctx, botID, telegramID, "active")
}

func (s *SubscriptionService) GetStatus(ctx context.Context, botID string, telegramID int64) (string, error) {
	sub, err := s.repo.Get(ctx, botID, telegramID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "none", nil
	}
	return sub.Status, nil
}

func (s *SubscriptionService) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	return s.repo.ListAll(ctx)
}
