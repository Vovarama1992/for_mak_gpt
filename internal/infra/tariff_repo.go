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
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
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
			&t.Code,
			&t.Name,
			&t.Price,
			&t.DurationMinutes,
			&t.VoiceMinutes,
			&t.Description,
		); err != nil {
			return nil, err
		}
		plans = append(plans, &t)
	}
	return plans, rows.Err()
}

func (r *tariffRepo) Create(ctx context.Context, plan *ports.TariffPlan) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO tariff_plans (
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			description
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			description
	`,
		plan.Code,
		plan.Name,
		plan.Price,
		plan.DurationMinutes,
		plan.VoiceMinutes,
		plan.Description,
	)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.Description,
	); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *tariffRepo) GetByID(ctx context.Context, id int) (*ports.TariffPlan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT 
			id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			description
		FROM tariff_plans
		WHERE id = $1
	`, id)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
		&t.Description,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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
			description = $6
		WHERE id = $7
		RETURNING
			id,
			code,
			name,
			price,
			duration_minutes,
			voice_minutes,
			description
	`,
		plan.Code,
		plan.Name,
		plan.Price,
		plan.DurationMinutes,
		plan.VoiceMinutes,
		plan.Description,
		plan.ID,
	)

	var t ports.TariffPlan
	if err := row.Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Price,
		&t.DurationMinutes,
		&t.VoiceMinutes,
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
