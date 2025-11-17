package ports

import (
	"context"
	"time"
)

type Subscription struct {
	ID                int64      `json:"id"`
	BotID             string     `json:"bot_id"`
	TelegramID        int64      `json:"telegram_id"`
	PlanID            int64      `json:"plan_id"`
	PlanName          string     `json:"plan_name,omitempty"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at"`
	ExpiresAt         *time.Time `json:"expires_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	YookassaPaymentID *string    `json:"yookassa_payment_id"`
	VoiceMinutes      int        `json:"voice_minutes"` // ← целочисленно
}

type SubscriptionRepo interface {
	Create(ctx context.Context, s *Subscription) error
	GetByPaymentID(ctx context.Context, paymentID string) (*Subscription, error)
	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	ListAll(ctx context.Context) ([]*Subscription, error)
	UseVoiceMinutes(ctx context.Context, botID string, tgID int64, used float64) (bool, error)

	// добавили voiceMinutes
	Activate(ctx context.Context, id int64, startedAt, expiresAt time.Time, voiceMinutes int) error
}
