package classes

import (
	"context"
	"database/sql"
)

type repo struct {
	db *sql.DB
}

func NewClassRepo(db *sql.DB) ClassRepo {
	return &repo{db: db}
}

//
// class_prompts
//

func (r *repo) CreatePrompt(ctx context.Context, p *ClassPrompt) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO class_prompts (class, prompt)
		 VALUES ($1, $2)
		 RETURNING id`,
		p.Class, p.Prompt,
	).Scan(&p.ID)
}

func (r *repo) UpdatePrompt(ctx context.Context, p *ClassPrompt) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE class_prompts
		 SET class=$1, prompt=$2
		 WHERE id=$3`,
		p.Class, p.Prompt, p.ID,
	)
	return err
}

func (r *repo) DeletePrompt(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM class_prompts WHERE id=$1`,
		id,
	)
	return err
}

func (r *repo) GetPromptByID(ctx context.Context, id int) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, class, prompt
		 FROM class_prompts
		 WHERE id=$1`,
		id,
	)

	var p ClassPrompt
	err := row.Scan(&p.ID, &p.Class, &p.Prompt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repo) GetPromptByClass(ctx context.Context, class int) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, class, prompt
		 FROM class_prompts
		 WHERE class=$1`,
		class,
	)

	var p ClassPrompt
	err := row.Scan(&p.ID, &p.Class, &p.Prompt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repo) ListPrompts(ctx context.Context) ([]*ClassPrompt, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, class, prompt
		 FROM class_prompts
		 ORDER BY class ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*ClassPrompt
	for rows.Next() {
		var p ClassPrompt
		if err := rows.Scan(&p.ID, &p.Class, &p.Prompt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

//
// user_classes
//

func (r *repo) SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_classes (bot_id, telegram_id, class_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (bot_id, telegram_id)
		 DO UPDATE SET class_id = EXCLUDED.class_id`,
		botID, telegramID, classID,
	)
	return err
}

func (r *repo) GetUserClass(ctx context.Context, botID string, telegramID int64) (*UserClass, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT bot_id, telegram_id, class_id
		 FROM user_classes
		 WHERE bot_id=$1 AND telegram_id=$2`,
		botID, telegramID,
	)

	var uc UserClass
	err := row.Scan(&uc.BotID, &uc.TelegramID, &uc.ClassID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *repo) DeleteUserClass(ctx context.Context, botID string, telegramID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_classes
		 WHERE bot_id=$1 AND telegram_id=$2`,
		botID, telegramID,
	)
	return err
}
