package bots

import (
	"context"
	"database/sql"
	"fmt"
)

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return &repo{db: db}
}

// ListAll
func (r *repo) ListAll(ctx context.Context) ([]*BotConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT bot_id, token, model,
		       text_style_prompt, voice_style_prompt,
		       voice_id
		FROM bot_configs
		ORDER BY bot_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*BotConfig
	for rows.Next() {
		var b BotConfig
		if err := rows.Scan(
			&b.BotID,
			&b.Token,
			&b.Model,
			&b.TextStylePrompt,
			&b.VoiceStylePrompt,
			&b.VoiceID,
		); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// Get
func (r *repo) Get(ctx context.Context, botID string) (*BotConfig, error) {
	var b BotConfig
	err := r.db.QueryRowContext(ctx, `
		SELECT bot_id, token, model,
		       text_style_prompt, voice_style_prompt,
		       voice_id
		FROM bot_configs
		WHERE bot_id = $1
	`, botID).Scan(
		&b.BotID,
		&b.Token,
		&b.Model,
		&b.TextStylePrompt,
		&b.VoiceStylePrompt,
		&b.VoiceID,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// Update
func (r *repo) Update(ctx context.Context, in *UpdateInput) (*BotConfig, error) {
	q := "UPDATE bot_configs SET "
	args := []any{}
	idx := 1

	if in.Model != nil {
		q += "model=$" + itoa(idx) + ","
		args = append(args, *in.Model)
		idx++
	}
	if in.TextStylePrompt != nil {
		q += "text_style_prompt=$" + itoa(idx) + ","
		args = append(args, *in.TextStylePrompt)
		idx++
	}
	if in.VoiceStylePrompt != nil {
		q += "voice_style_prompt=$" + itoa(idx) + ","
		args = append(args, *in.VoiceStylePrompt)
		idx++
	}
	if in.VoiceID != nil {
		q += "voice_id=$" + itoa(idx) + ","
		args = append(args, *in.VoiceID)
		idx++
	}

	if len(args) == 0 {
		return r.Get(ctx, in.BotID)
	}

	q = q[:len(q)-1] // убрать последнюю запятую
	q += " WHERE bot_id=$" + itoa(idx) +
		" RETURNING bot_id, token, model, text_style_prompt, voice_style_prompt, voice_id"

	args = append(args, in.BotID)

	var b BotConfig
	err := r.db.QueryRowContext(ctx, q, args...).Scan(
		&b.BotID,
		&b.Token,
		&b.Model,
		&b.TextStylePrompt,
		&b.VoiceStylePrompt,
		&b.VoiceID,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
