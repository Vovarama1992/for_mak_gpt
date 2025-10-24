package domain

import (
	"context"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type subscriptionService struct {
	repo ports.SubscriptionRepo
}

func NewSubscriptionService(repo ports.SubscriptionRepo) ports.SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) Create(ctx context.Context, botID string, telegramID int64, planCode string) error {
	// на этом этапе просто создаём запись с pending
	sub := &ports.Subscription{
		BotID:      botID,
		TelegramID: telegramID,
		Status:     "pending",
		StartedAt:  time.Now(),
	}
	return s.repo.Create(ctx, sub)
}

func (s *subscriptionService) Activate(ctx context.Context, botID string, telegramID int64) error {
	return s.repo.UpdateStatus(ctx, botID, telegramID, "active")
}

func (s *subscriptionService) GetStatus(ctx context.Context, botID string, telegramID int64) (string, error) {
	sub, err := s.repo.Get(ctx, botID, telegramID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "none", nil
	}
	return sub.Status, nil
}

func (s *subscriptionService) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	return s.repo.ListAll(ctx)
}
