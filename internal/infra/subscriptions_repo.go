package infra

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type subscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) ports.SubscriptionRepo {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) scanRow(row *sql.Row) (*ports.Subscription, error) {
	var s ports.Subscription
	var yid sql.NullString

	err := row.Scan(
		&s.ID,
		&s.BotID,
		&s.TelegramID,
		&s.PlanID,
		&s.Status,
		&s.StartedAt,
		&s.ExpiresAt,
		&s.UpdatedAt,
		&yid,
		&s.VoiceMinutes, // ⬅️ ЯДРО ФИКСА
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if yid.Valid {
		s.YookassaPaymentID = &yid.String
	}

	return &s, nil
}

func (r *subscriptionRepo) Create(ctx context.Context, s *ports.Subscription) error {
	const q = `
		INSERT INTO subscriptions (
			bot_id, telegram_id, plan_id, status,
			started_at, expires_at, updated_at, yookassa_payment_id
		)
		VALUES ($1,$2,$3,$4,$5,$6, now(), $7)
		ON CONFLICT (bot_id, telegram_id)
		DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			expires_at = EXCLUDED.expires_at,
			updated_at = now(),
			yookassa_payment_id = EXCLUDED.yookassa_payment_id
		RETURNING id
	`
	return r.db.QueryRowContext(
		ctx, q,
		s.BotID,
		s.TelegramID,
		s.PlanID,
		s.Status,
		s.StartedAt,
		s.ExpiresAt,
		s.YookassaPaymentID,
	).Scan(&s.ID)
}

func (r *subscriptionRepo) Delete(ctx context.Context, botID string, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM subscriptions
        WHERE bot_id = $1 AND telegram_id = $2
    `, botID, telegramID)
	return err
}

func (r *subscriptionRepo) GetByPaymentID(ctx context.Context, paymentID string) (*ports.Subscription, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT 
			id, bot_id, telegram_id, plan_id, status,
			started_at, expires_at, updated_at, yookassa_payment_id,
			voice_minutes
		FROM subscriptions 
		WHERE yookassa_payment_id = $1
	`, paymentID)

	return r.scanRow(row)
}

func (r *subscriptionRepo) Get(
	ctx context.Context,
	botID string,
	telegramID int64,
) (*ports.Subscription, error) {

	row := r.db.QueryRowContext(ctx, `
		SELECT 
			id, bot_id, telegram_id, plan_id, status,
			started_at, expires_at, updated_at, yookassa_payment_id,
			voice_minutes
		FROM subscriptions
		WHERE bot_id = $1 AND telegram_id = $2
		ORDER BY
			CASE
				WHEN status = 'active' THEN 0
				ELSE 1
			END,
			started_at DESC
		LIMIT 1
	`, botID, telegramID)

	sub, err := r.scanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return sub, nil
}

func (r *subscriptionRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET status=$1, updated_at=$2 
		WHERE id=$3
	`, status, time.Now(), id)

	return err
}

func (r *subscriptionRepo) Activate(
	ctx context.Context,
	id int64,
	startedAt, expiresAt time.Time,
	voiceMinutes float64,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET 
			status        = 'active',
			started_at    = $2,
			expires_at    = $3,
			voice_minutes = $4,
			updated_at    = NOW()
		WHERE id = $1
	`, id, startedAt, expiresAt, voiceMinutes)

	return err
}

func (r *subscriptionRepo) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			s.id,
			s.bot_id,
			s.telegram_id,
			s.plan_id,
			s.status,
			s.started_at,
			s.expires_at,
			s.updated_at,
			s.yookassa_payment_id,
			tp.name AS plan_name,
			s.voice_minutes         -- ⬅️ ТУТ ДОБАВЛЕНО
		FROM subscriptions s
		LEFT JOIN tariff_plans tp ON tp.id = s.plan_id
		ORDER BY s.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*ports.Subscription

	for rows.Next() {
		var s ports.Subscription
		var yid sql.NullString
		var planName sql.NullString

		err := rows.Scan(
			&s.ID,
			&s.BotID,
			&s.TelegramID,
			&s.PlanID,
			&s.Status,
			&s.StartedAt,
			&s.ExpiresAt,
			&s.UpdatedAt,
			&yid,
			&planName,
			&s.VoiceMinutes,
		)
		if err != nil {
			return nil, err
		}

		if yid.Valid {
			s.YookassaPaymentID = &yid.String
		}
		if planName.Valid {
			s.PlanName = planName.String
		}

		subs = append(subs, &s)
	}

	return subs, rows.Err()
}

func (r *subscriptionRepo) UseVoiceMinutes(ctx context.Context, botID string, tgID int64, used float64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET voice_minutes = voice_minutes - $3
		WHERE bot_id=$1 AND telegram_id=$2 AND voice_minutes >= $3
	`, botID, tgID, used)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *subscriptionRepo) AddVoiceMinutes(
	ctx context.Context,
	botID string,
	tgID int64,
	minutes float64,
) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE subscriptions
        SET voice_minutes = voice_minutes + $3,
            updated_at = NOW()
        WHERE bot_id = $1 AND telegram_id = $2
    `, botID, tgID, minutes)

	return err
}

func (r *subscriptionRepo) CleanupPending(ctx context.Context, olderThan time.Duration) error {
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM subscriptions
        WHERE status = 'pending'
        AND updated_at < NOW() - $1::interval
    `, olderThan.String())
	return err
}

func (r *subscriptionRepo) CreateDemo(
	ctx context.Context,
	botID string,
	telegramID int64,
	startedAt, expiresAt time.Time,
	voiceMinutes float64,
) error {

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO subscriptions (
			bot_id,
			telegram_id,
			plan_id,
			status,
			started_at,
			expires_at,
			updated_at,
			yookassa_payment_id,
			voice_minutes
		)
		VALUES (
			$1, $2, NULL, 'active', $3, $4, NOW(), NULL, $5
		)
		ON CONFLICT (bot_id, telegram_id)
		DO UPDATE SET
			status = 'active',
			plan_id = NULL,
			started_at = EXCLUDED.started_at,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW(),
			yookassa_payment_id = NULL,
			voice_minutes = EXCLUDED.voice_minutes;
	`, botID, telegramID, startedAt, expiresAt, voiceMinutes)

	return err
}
