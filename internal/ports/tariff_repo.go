package ports

import (
	"context"
	"encoding/json"
)

type TariffRepo interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
	GetByID(ctx context.Context, botID string, id int) (*TariffPlan, error)
	GetTrial(ctx context.Context, botID string) (*TariffPlan, error)
	Create(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
	Update(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
	Delete(ctx context.Context, id int) error
}

type TariffService interface {
	ListAll(ctx context.Context) ([]*TariffPlan, error)
	GetByID(ctx context.Context, botID string, id int) (*TariffPlan, error)
	GetTrial(ctx context.Context, botID string) (*TariffPlan, error)
	Create(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
	Update(ctx context.Context, plan *TariffPlan) (*TariffPlan, error)
	Delete(ctx context.Context, id int) error
}

type TariffPlan struct {
	ID    int     `json:"id"`
	BotID string  `json:"bot_id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`

	DurationMinutes int     `json:"duration_minutes"`
	VoiceMinutes    float64 `json:"voice_minutes"`

	IsTrial bool `json:"is_trial"`

	Description string          `json:"description"`
	Features    json.RawMessage `json:"features"`
}
