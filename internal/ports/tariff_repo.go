package ports

import (
	"context"
	"encoding/json"
)

type TariffRepo interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
}

type TariffService interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
}

type TariffPlan struct {
	ID           int             `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Price        float64         `json:"price"`
	PeriodDays   int             `json:"period_days"`
	VoiceMinutes int             `json:"voice_minutes"`
	Description  string          `json:"description"`
	Features     json.RawMessage `json:"features"`
}
