package domain

import (
	"context"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type TariffService struct {
	repo ports.TariffRepo
}

func NewTariffService(repo ports.TariffRepo) ports.TariffService {
	return &TariffService{repo: repo}
}

func (s *TariffService) ListAll(ctx context.Context, botID string) ([]*ports.TariffPlan, error) {
	return s.repo.ListAll(ctx, botID)
}

func (s *TariffService) GetByID(ctx context.Context, botID string, id int) (*ports.TariffPlan, error) {
	return s.repo.GetByID(ctx, botID, id)
}

func (s *TariffService) GetTrial(ctx context.Context, botID string) (*ports.TariffPlan, error) {
	return s.repo.GetTrial(ctx, botID)
}

func (s *TariffService) Create(ctx context.Context, plan *ports.TariffPlan) (*ports.TariffPlan, error) {
	return s.repo.Create(ctx, plan)
}

func (s *TariffService) Update(ctx context.Context, plan *ports.TariffPlan) (*ports.TariffPlan, error) {
	return s.repo.Update(ctx, plan)
}

func (s *TariffService) Delete(ctx context.Context, botID string, id int) error {
	return s.repo.Delete(ctx, botID, id)
}
