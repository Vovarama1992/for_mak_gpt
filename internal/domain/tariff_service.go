package domain

import (
	"context"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type tariffService struct {
	repo ports.TariffRepo
}

func NewTariffService(repo ports.TariffRepo) ports.TariffService {
	return &tariffService{repo: repo}
}

func (s *tariffService) ListAll(ctx context.Context) ([]*ports.TariffPlan, error) {
	return s.repo.ListAll(ctx)
}
