package infra

import (
	"context"
	"database/sql"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) GetPassword(ctx context.Context) (string, error) {
	var password string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT password FROM bot_auth LIMIT 1`,
	).Scan(&password)

	if err == sql.ErrNoRows {
		return "", nil
	}
	return password, err
}
