package ports

import (
	"context"
	"encoding/json"
)

type TariffRepo interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
	GetByID(ctx context.Context, id int) (*TariffPlan, error)

	Create(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
}

type TariffService interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
	GetByID(ctx context.Context, id int) (*TariffPlan, error)

	Create(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
}

type TariffPlan struct {
	ID              int             `json:"id"`
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Price           float64         `json:"price"`
	DurationMinutes int             `json:"duration_minutes"` // ← единая ось времени
	VoiceMinutes    float64         `json:"voice_minutes"`
	Description     string          `json:"description"`
	Features        json.RawMessage `json:"features"`
}
