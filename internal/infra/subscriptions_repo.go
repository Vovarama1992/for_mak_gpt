package infra

import (
	"context"
	"database/sql"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type subscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) ports.SubscriptionRepo {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, s *ports.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			bot_id,
			telegram_id,
			plan_id,
			status,
			started_at,
			expires_at,
			updated_at,
			yookassa_payment_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		s.BotID,
		s.TelegramID,
		s.PlanID,
		s.Status,
		s.StartedAt,
		s.ExpiresAt,
		time.Now(),
		s.YookassaPaymentID,
	).Scan(&s.ID)
}

func (r *subscriptionRepo) GetByPaymentID(ctx context.Context, paymentID string) (*ports.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at, yookassa_payment_id
		FROM subscriptions
		WHERE yookassa_payment_id = $1
	`, paymentID)

	var s ports.Subscription
	err := row.Scan(
		&s.ID,
		&s.BotID,
		&s.TelegramID,
		&s.PlanID,
		&s.Status,
		&s.StartedAt,
		&s.ExpiresAt,
		&s.UpdatedAt,
		&s.YookassaPaymentID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *subscriptionRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, status, time.Now(), id)
	return err
}

func (r *subscriptionRepo) Get(ctx context.Context, botID string, telegramID int64) (*ports.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at, yookassa_payment_id
		FROM subscriptions
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID)

	var s ports.Subscription
	err := row.Scan(
		&s.ID,
		&s.BotID,
		&s.TelegramID,
		&s.PlanID,
		&s.Status,
		&s.StartedAt,
		&s.ExpiresAt,
		&s.UpdatedAt,
		&s.YookassaPaymentID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *subscriptionRepo) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at, yookassa_payment_id
		FROM subscriptions
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*ports.Subscription
	for rows.Next() {
		var s ports.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.BotID,
			&s.TelegramID,
			&s.PlanID,
			&s.Status,
			&s.StartedAt,
			&s.ExpiresAt,
			&s.UpdatedAt,
			&s.YookassaPaymentID,
		); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, rows.Err()
}
