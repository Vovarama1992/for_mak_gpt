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
		INSERT INTO subscriptions (bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
	).Scan(&s.ID)
}

func (r *subscriptionRepo) UpdateStatus(ctx context.Context, botID string, telegramID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET status = $1, updated_at = $2
		WHERE bot_id = $3 AND telegram_id = $4
	`, status, time.Now(), botID, telegramID)
	return err
}

func (r *subscriptionRepo) Get(ctx context.Context, botID string, telegramID int64) (*ports.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at
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
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *subscriptionRepo) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, bot_id, telegram_id, plan_id, status, started_at, expires_at, updated_at
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
		); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, rows.Err()
}
