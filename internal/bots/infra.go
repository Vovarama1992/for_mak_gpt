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

func (r *repo) Create(ctx context.Context, in *CreateInput) (*BotConfig, error) {
	var b BotConfig

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO bot_configs (
			bot_id,
			token,
			model,
			voice_id
		) VALUES ($1, $2, $3, $4)
		RETURNING
			bot_id,
			token,
			model,
			text_style_prompt,
			voice_style_prompt,
			photo_style_prompt,
			voice_id,
			welcome_text,
			tariff_text,
			after_continue_text,
			no_voice_minutes_text,
			welcome_video
	`,
		in.BotID,
		in.Token,
		in.Model,
		in.VoiceID,
	).Scan(
		&b.BotID,
		&b.Token,
		&b.Model,
		&b.TextStylePrompt,
		&b.VoiceStylePrompt,
		&b.PhotoStylePrompt,
		&b.VoiceID,
		&b.WelcomeText,
		&b.TariffText,
		&b.AfterContinueText,
		&b.NoVoiceMinutesText,
		&b.WelcomeVideo,
	)

	if err != nil {
		return nil, err
	}

	return &b, nil
}

// --------------------------------------------------
// ListAll
// --------------------------------------------------

func (r *repo) ListAll(ctx context.Context) ([]*BotConfig, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT bot_id, token, model,
		       text_style_prompt,
		       voice_style_prompt,
		       photo_style_prompt,
		       voice_id,
		       welcome_text,
		       tariff_text,
		       after_continue_text,
		       no_voice_minutes_text,
		       welcome_video
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
			&b.PhotoStylePrompt,
			&b.VoiceID,
			&b.WelcomeText,
			&b.TariffText,
			&b.AfterContinueText,
			&b.NoVoiceMinutesText,
			&b.WelcomeVideo,
		); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}

	return out, rows.Err()
}

// --------------------------------------------------
// Get
// --------------------------------------------------

func (r *repo) Get(ctx context.Context, botID string) (*BotConfig, error) {
	var b BotConfig

	err := r.db.QueryRowContext(ctx, `
		SELECT bot_id, token, model,
		       text_style_prompt,
		       voice_style_prompt,
		       photo_style_prompt,
		       voice_id,
		       welcome_text,
		       tariff_text,
		       after_continue_text,
		       no_voice_minutes_text,
		       welcome_video
		FROM bot_configs
		WHERE bot_id = $1
	`, botID).Scan(
		&b.BotID,
		&b.Token,
		&b.Model,
		&b.TextStylePrompt,
		&b.VoiceStylePrompt,
		&b.PhotoStylePrompt,
		&b.VoiceID,
		&b.WelcomeText,
		&b.TariffText,
		&b.AfterContinueText,
		&b.NoVoiceMinutesText,
		&b.WelcomeVideo,
	)

	if err != nil {
		return nil, err
	}

	return &b, nil
}

// --------------------------------------------------
// Update
// --------------------------------------------------

func (r *repo) Update(ctx context.Context, in *UpdateInput) (*BotConfig, error) {
	q := "UPDATE bot_configs SET "
	args := []any{}
	idx := 1

	appendField := func(field string, value *string) {
		if value != nil {
			q += field + "=$" + itoa(idx) + ","
			args = append(args, *value)
			idx++
		}
	}

	// теперь bot_id — обычное поле
	appendField("bot_id", in.NewBotID)

	appendField("token", in.Token)
	appendField("model", in.Model)
	appendField("text_style_prompt", in.TextStylePrompt)
	appendField("voice_style_prompt", in.VoiceStylePrompt)
	appendField("photo_style_prompt", in.PhotoStylePrompt)
	appendField("voice_id", in.VoiceID)
	appendField("welcome_text", in.WelcomeText)
	appendField("tariff_text", in.TariffText)
	appendField("after_continue_text", in.AfterContinueText)
	appendField("no_voice_minutes_text", in.NoVoiceMinutesText)
	appendField("welcome_video", in.WelcomeVideo)

	if len(args) == 0 {
		return r.Get(ctx, in.BotID)
	}

	q = q[:len(q)-1]

	q += `
		WHERE bot_id = $` + itoa(idx) + `
		RETURNING
			bot_id,
			token,
			model,
			text_style_prompt,
			voice_style_prompt,
			photo_style_prompt,
			voice_id,
			welcome_text,
			tariff_text,
			after_continue_text,
			no_voice_minutes_text,
			welcome_video
	`

	args = append(args, in.BotID)

	var b BotConfig
	err := r.db.QueryRowContext(ctx, q, args...).Scan(
		&b.BotID,
		&b.Token,
		&b.Model,
		&b.TextStylePrompt,
		&b.VoiceStylePrompt,
		&b.PhotoStylePrompt,
		&b.VoiceID,
		&b.WelcomeText,
		&b.TariffText,
		&b.AfterContinueText,
		&b.NoVoiceMinutesText,
		&b.WelcomeVideo,
	)
	if err != nil {
		return nil, fmt.Errorf("update bot_configs failed: %w | SQL=%s ARGS=%v", err, q, args)
	}

	return &b, nil
}

func (r *repo) Delete(ctx context.Context, botID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM bot_configs WHERE bot_id = $1`,
		botID,
	)
	return err
}

// --------------------------------------------------

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
