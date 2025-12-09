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

func (r *repo) CreateClass(ctx context.Context, grade string) (*Class, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO classes (grade)
		 VALUES ($1)
		 RETURNING id, grade`,
		grade,
	)
	var c Class
	if err := row.Scan(&c.ID, &c.Grade); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repo) ListClasses(ctx context.Context) ([]*Class, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			c.id,
			c.grade,
			p.id   AS prompt_id,
			p.prompt
		FROM classes c
		LEFT JOIN class_prompts p ON p.class_id = c.id
		ORDER BY c.grade ASC
	`)
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

		if err := rows.Scan(&c.ID, &c.Grade, &pid, &pp); err != nil {
			return nil, err
		}

		// если промпт есть
		if pid.Valid {
			c.Prompt = &ClassPrompt{
				ID:      int(pid.Int64),
				ClassID: c.ID,
				Prompt:  pp.String,
			}
		}

		out = append(out, &c)
	}
	return out, rows.Err()
}

func (r *repo) UpdateClass(ctx context.Context, id int, grade string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE classes
	     SET grade=$1
	     WHERE id=$2`,
		grade, id,
	)
	return err
}

// удалить класс
func (r *repo) DeleteClass(ctx context.Context, id int) error {
	// 1) удалить всех юзеров, у которых этот класс
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_classes WHERE class_id=$1`,
		id,
	)
	if err != nil {
		return err
	}

	// 2) удалить сам класс
	_, err = r.db.ExecContext(ctx,
		`DELETE FROM classes WHERE id=$1`,
		id,
	)
	return err
}

func (r *repo) GetClassByID(ctx context.Context, id int) (*Class, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, grade
		 FROM classes
		 WHERE id=$1`,
		id,
	)

	var c Class
	if err := row.Scan(&c.ID, &c.Grade); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

//
// class_prompts
//

func (r *repo) CreatePrompt(ctx context.Context, classID int, prompt string) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO class_prompts (class_id, prompt)
		 VALUES ($1, $2)
		 RETURNING id, class_id, prompt`,
		classID, prompt,
	)

	var p ClassPrompt
	if err := row.Scan(&p.ID, &p.ClassID, &p.Prompt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repo) UpdatePrompt(ctx context.Context, id int, prompt string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE class_prompts
		 SET prompt=$1
		 WHERE id=$2`,
		prompt, id,
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

func (r *repo) GetPromptByClassID(ctx context.Context, classID int) (*ClassPrompt, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, class_id, prompt
		 FROM class_prompts
		 WHERE class_id=$1`,
		classID,
	)

	var p ClassPrompt
	if err := row.Scan(&p.ID, &p.ClassID, &p.Prompt); err != nil {
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
