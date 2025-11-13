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
			description,
			features
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
			&t.Features,
		); err != nil {
			return nil, err
		}
		plans = append(plans, &t)
	}
	return plans, rows.Err()
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
			description,
			features
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
		&t.Features,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}
