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

	// 2) подписка
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM subscriptions
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID); err != nil {
		return err
	}

	// 3) состояние истории (курсор / токены)
	if _, err := tx.ExecContext(ctx, `
		UPDATE history_state
		SET last_n_records = 0,
		    total_tokens = 0,
		    updated_at = NOW()
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID); err != nil {
		return err
	}

	return tx.Commit()
}
