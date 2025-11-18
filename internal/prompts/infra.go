package prompts

import (
	"context"
	"database/sql"
)

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return &repo{db: db}
}

func (r *repo) ListAll(ctx context.Context) ([]*Prompt, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT bot_id, prompt
		FROM bot_prompts
		ORDER BY bot_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Prompt
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.BotID, &p.Prompt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (r *repo) Update(ctx context.Context, botID, prompt string) (*Prompt, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE bot_prompts
		SET prompt = $1
		WHERE bot_id = $2
		RETURNING bot_id, prompt
	`, prompt, botID)

	var p Prompt
	if err := row.Scan(&p.BotID, &p.Prompt); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *repo) GetByBotID(ctx context.Context, botID string) (string, error) {
	var prompt string
	err := r.db.QueryRowContext(ctx, `
		SELECT prompt
		FROM bot_prompts
		WHERE bot_id = $1
	`, botID).Scan(&prompt)
	if err != nil {
		return "", err
	}
	return prompt, nil
}
