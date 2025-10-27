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

func (s *TariffService) ListAll(ctx context.Context) ([]*ports.TariffPlan, error) {
	return s.repo.ListAll(ctx)
}
