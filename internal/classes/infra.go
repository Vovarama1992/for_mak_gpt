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
// classes
//

func (r *repo) CreateClass(ctx context.Context, botID string, grade string) (*Class, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO classes (bot_id, bot_ref, grade)
		 VALUES (
			$1,
			(SELECT id FROM bot_configs WHERE bot_id = $1),
			$2
		 )
		 RETURNING id, bot_id, grade`,
		botID, grade,
	)

	var c Class
	if err := row.Scan(&c.ID, &c.BotID, &c.Grade); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repo) ListClasses(ctx context.Context, botID string) ([]*Class, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			c.id,
			c.bot_id,
			c.grade,
			p.id   AS prompt_id,
			p.prompt
		FROM classes c
		LEFT JOIN class_prompts p 
			ON p.class_id = c.id 
			AND p.bot_ref = c.bot_ref
		WHERE c.bot_id = $1
		ORDER BY c.grade ASC
	`, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Class

	for rows.Next() {
		var (
			c   Class
			pid sql.NullInt64
			pp  sql.NullString
		)

		if err := rows.Scan(
			&c.ID,
			&c.BotID,
			&c.Grade,
			&pid,
			&pp,
		); err != nil {
			return nil, err
		}

		if pid.Valid {
			c.Prompt = &ClassPrompt{
				ID:      int(pid.Int64),
				BotID:   c.BotID,
				ClassID: c.ID,
				Prompt:  pp.String,
			}
		}

		out = append(out, &c)
	}

	return out, rows.Err()
}

func (r *repo) GetClassByID(ctx context.Context, botID string, id int) (*Class, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, bot_id, grade
		 FROM classes
		 WHERE id=$1 AND bot_id=$2`,
		id, botID,
	)

	var c Class
	if err := row.Scan(&c.ID, &c.BotID, &c.Grade); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *repo) UpdateClass(ctx context.Context, botID string, id int, grade string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE classes
		 SET grade=$1
		 WHERE id=$2 AND bot_id=$3`,
		grade, id, botID,
	)
	return err
}

func (r *repo) DeleteClassByID(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_classes
		 WHERE class_id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`DELETE FROM class_prompts
		 WHERE class_id = $1`,
		id,
	)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`DELETE FROM classes
		 WHERE id = $1`,
		id,
	)

	return err
}

//
// class_prompts
//

func (r *repo) CreatePrompt(ctx context.Context, botID string, classID int, prompt string) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO class_prompts (bot_id, bot_ref, class_id, prompt)
		 VALUES (
			$1,
			(SELECT id FROM bot_configs WHERE bot_id = $1),
			$2,
			$3
		 )
		 RETURNING id, bot_id, class_id, prompt`,
		botID, classID, prompt,
	)

	var p ClassPrompt
	if err := row.Scan(&p.ID, &p.BotID, &p.ClassID, &p.Prompt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repo) UpdatePrompt(ctx context.Context, botID string, id int, prompt string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE class_prompts
		 SET prompt=$1
		 WHERE id=$2 AND bot_id=$3`,
		prompt, id, botID,
	)
	return err
}

func (r *repo) DeletePrompt(ctx context.Context, botID string, id int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM class_prompts
		 WHERE id=$1 AND bot_id=$2`,
		id, botID,
	)
	return err
}

func (r *repo) GetPromptByClassID(ctx context.Context, botID string, classID int) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, bot_id, class_id, prompt
		 FROM class_prompts
		 WHERE class_id=$1 AND bot_id=$2`,
		classID, botID,
	)

	var p ClassPrompt
	if err := row.Scan(&p.ID, &p.BotID, &p.ClassID, &p.Prompt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

//
// user_classes
//

func (r *repo) SetUserClass(ctx context.Context, botID string, telegramID int64, classID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_classes (bot_id, bot_ref, telegram_id, class_id)
		 VALUES (
			$1,
			(SELECT id FROM bot_configs WHERE bot_id = $1),
			$2,
			$3
		 )
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
	if err := row.Scan(&uc.BotID, &uc.TelegramID, &uc.ClassID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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
