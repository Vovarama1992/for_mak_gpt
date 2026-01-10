package trial

import (
	"context"
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) RepoInf {
	return &Repo{db: db}
}

// Exists — проверка наличия записи
func (r *Repo) Exists(
	ctx context.Context,
	botID string,
	telegramID int64,
) (bool, error) {

	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM trial_usages
			WHERE bot_id = $1 AND telegram_id = $2
		)
	`, botID, telegramID).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// Create — фиксируем факт trial
func (r *Repo) Create(
	ctx context.Context,
	botID string,
	telegramID int64,
) error {

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO trial_usages (bot_id, telegram_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, botID, telegramID)

	return err
}

// Delete — удаляем факт использования trial
func (r *Repo) Delete(
	ctx context.Context,
	botID string,
	telegramID int64,
) error {

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM trial_usages
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID)

	return err
}
