package infra

import (
	"context"
	"database/sql"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type tariffRepo struct {
	db *sql.DB
}

func NewTariffRepo(db *sql.DB) ports.TariffRepo {
	return &tariffRepo{db: db}
}

func (r *tariffRepo) ListAll(ctx context.Context) ([]*ports.TariffPlan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
		FROM tariff_plans
		ORDER BY price ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*ports.TariffPlan

	for rows.Next() {
		var t ports.TariffPlan
		if err := rows.Scan(
			&t.ID,
			&t.BotID,
			&t.Code,
			&t.Name,
			&t.Price,
			&t.DurationMinutes,
			&t.VoiceMinutes,
			&t.IsTrial,
			&t.Description,
		); err != nil {
			return nil, err
		}
		plans = append(plans, &t)
	}

	return plans, rows.Err()
}

func (r *tariffRepo) GetByID(ctx context.Context, botID string, id int) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
		FROM tariff_plans
		WHERE id = $1 AND bot_id = $2
	`, id, botID)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.BotID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.IsTrial,
		&t.Description,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

func (r *tariffRepo) GetTrial(ctx context.Context, botID string) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
		FROM tariff_plans
		WHERE bot_id = $1 AND is_trial = true
		LIMIT 1
	`, botID)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.BotID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.IsTrial,
		&t.Description,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

func (r *tariffRepo) Create(ctx context.Context, plan *ports.TariffPlan) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO tariff_plans (
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id,
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
	`,
		plan.BotID,
		plan.Code,
		plan.Name,
		plan.Price,
		plan.DurationMinutes,
		plan.VoiceMinutes,
		plan.IsTrial,
		plan.Description,
	)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.BotID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.IsTrial,
		&t.Description,
	); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *tariffRepo) Update(ctx context.Context, plan *ports.TariffPlan) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE tariff_plans
		SET
			code = $1,
			name = $2,
			price = $3,
			duration_minutes = $4,
			voice_minutes = $5,
			is_trial = $6,
			description = $7
		WHERE id = $8 AND bot_id = $9
		RETURNING
			id,
			bot_id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			is_trial,
			description
	`,
		plan.Code,
		plan.Name,
		plan.Price,
		plan.DurationMinutes,
		plan.VoiceMinutes,
		plan.IsTrial,
		plan.Description,
		plan.ID,
		plan.BotID,
	)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.BotID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.IsTrial,
		&t.Description,
	); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *tariffRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM tariff_plans
		WHERE id = $1
	`, id)
	return err
}
