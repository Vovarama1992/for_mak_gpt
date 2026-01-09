package user

import (
	"context"
	"database/sql"
)

type infra struct {
	db *sql.DB
}

func NewInfra(db *sql.DB) Infra {
	return &infra{db: db}
}

func (i *infra) ResetUserSettings(
	ctx context.Context,
	botID string,
	telegramID int64,
) error {

	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1) выбранный класс
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_classes
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID); err != nil {
		return err
	}

	return tx.Commit()
}
