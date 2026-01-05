package ports

import (
	"context"
	"time"
)

type Subscription struct {
	ID                int64      `json:"id"`
	BotID             string     `json:"bot_id"`
	TelegramID        int64      `json:"telegram_id"`
	PlanID            *int64     `json:"plan_id"`
	PlanName          string     `json:"plan_name,omitempty"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at"`
	ExpiresAt         *time.Time `json:"expires_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	TrialNotifiedAt   *time.Time `json:"trial_notified_at"` // ← ДОБАВИЛИ
	YookassaPaymentID *string    `json:"yookassa_payment_id"`
	VoiceMinutes      float64    `json:"voice_minutes"`
}

type SubscriptionRepo interface {
	Create(ctx context.Context, s *Subscription) error
	GetByPaymentID(ctx context.Context, paymentID string) (*Subscription, error)
	Get(ctx context.Context, botID string, telegramID int64) (*Subscription, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	ListAll(ctx context.Context) ([]*Subscription, error)
	UseVoiceMinutes(ctx context.Context, botID string, tgID int64, used float64) (bool, error)
	AddVoiceMinutes(ctx context.Context, botID string, tgID int64, minutes float64) error

	Delete(ctx context.Context, botID string, telegramID int64) error
	GetExpiredTrialsForNotify(ctx context.Context) ([]*Subscription, error)
	MarkTrialNotified(ctx context.Context, id int64) error

	CleanupPending(ctx context.Context, olderThan time.Duration) error
	Activate(ctx context.Context, id int64, startedAt, expiresAt time.Time, voiceMinutes float64) error
	CreateDemo(ctx context.Context, botID string, telegramID int64, startedAt, expiresAt time.Time, voiceMinutes float64) error
	UpdateLimits(
		ctx context.Context,
		id int64,
		expiresAt time.Time,
		voiceMinutes float64,
		status string,
	) error
}
