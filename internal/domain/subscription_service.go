package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type SubscriptionService struct {
	repo       ports.SubscriptionRepo
	tariffRepo ports.TariffRepo
}

func NewSubscriptionService(repo ports.SubscriptionRepo, tariffRepo ports.TariffRepo) ports.SubscriptionService {
	return &SubscriptionService{repo: repo, tariffRepo: tariffRepo}
}

func (s *SubscriptionService) Create(ctx context.Context, botID string, telegramID int64, planCode string) error {
	tariffs, err := s.tariffRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tariffs: %w", err)
	}

	var planID int
	for _, t := range tariffs {
		if t.Code == planCode {
			planID = t.ID
			break
		}
	}
	if planID == 0 {
		return fmt.Errorf("unknown plan code: %s", planCode)
	}

	sub := &ports.Subscription{
		BotID:      botID,
		TelegramID: telegramID,
		PlanID:     planID,
		Status:     "pending",
		StartedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
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
