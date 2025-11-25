package minutes_packages

import (
	"context"
	"database/sql"
)

type repo struct {
	db *sql.DB
}

func NewMinutePackageRepo(db *sql.DB) MinutePackageRepo {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, pkg *MinutePackage) error {
	query := `
		INSERT INTO minute_packages (name, minutes, price, active)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	return r.db.QueryRowContext(
		ctx, query,
		pkg.Name,
		pkg.Minutes,
		pkg.Price,
		pkg.Active,
	).Scan(&pkg.ID)
}

func (r *repo) Update(ctx context.Context, pkg *MinutePackage) error {
	query := `
		UPDATE minute_packages
		SET name=$1, minutes=$2, price=$3, active=$4
		WHERE id=$5
	`
	_, err := r.db.ExecContext(
		ctx, query,
		pkg.Name,
		pkg.Minutes,
		pkg.Price,
		pkg.Active,
		pkg.ID,
	)
	return err
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM minute_packages WHERE id=$1`,
		id,
	)
	return err
}

func (r *repo) GetByID(ctx context.Context, id int64) (*MinutePackage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, minutes, price, active
		FROM minute_packages
		WHERE id=$1
	`, id)

	var pkg MinutePackage
	err := row.Scan(
		&pkg.ID,
		&pkg.Name,
		&pkg.Minutes,
		&pkg.Price,
		&pkg.Active,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &pkg, nil
}

func (r *repo) ListAll(ctx context.Context) ([]*MinutePackage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, minutes, price, active
		FROM minute_packages
		ORDER BY minutes ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*MinutePackage

	for rows.Next() {
		var pkg MinutePackage

		err := rows.Scan(
			&pkg.ID,
			&pkg.Name,
			&pkg.Minutes,
			&pkg.Price,
			&pkg.Active,
		)
		if err != nil {
			return nil, err
		}

		out = append(out, &pkg)
	}

	return out, rows.Err()
}
