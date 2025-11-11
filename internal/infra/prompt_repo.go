package infra

import (
	"context"
	"database/sql"
)

type PromptRepo struct {
	db *sql.DB
}

func NewPromptRepo(db *sql.DB) *PromptRepo {
	return &PromptRepo{db: db}
}

func (r *PromptRepo) GetByBotID(ctx context.Context, botID string) (string, error) {
	var prompt string
	err := r.db.QueryRowContext(ctx, `SELECT prompt FROM bot_prompts WHERE bot_id = $1`, botID).Scan(&prompt)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return prompt, err
}