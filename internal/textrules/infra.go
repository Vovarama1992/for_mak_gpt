package textrules

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

// ===== LETTERS =====

func (r *repo) ListLetterRules(ctx context.Context) ([]LetterRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT from_char, to_char FROM text_letter_rules`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LetterRule
	for rows.Next() {
		var rr LetterRule
		if err := rows.Scan(&rr.From, &rr.To); err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

func (r *repo) AddLetterRule(ctx context.Context, from, to string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO text_letter_rules (from_char, to_char)
		 VALUES ($1, $2)
		 ON CONFLICT (from_char) DO UPDATE SET to_char = EXCLUDED.to_char`,
		from, to,
	)
	return err
}

func (r *repo) DeleteLetterRule(ctx context.Context, from string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM text_letter_rules WHERE from_char = $1`,
		from,
	)
	return err
}

// ===== WORDS =====

func (r *repo) ListWordRules(ctx context.Context) ([]WordRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT from_word, to_word FROM text_word_rules`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []WordRule
	for rows.Next() {
		var rr WordRule
		if err := rows.Scan(&rr.From, &rr.To); err != nil {
			return nil, err
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

func (r *repo) AddWordRule(ctx context.Context, from, to string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO text_word_rules (from_word, to_word)
		 VALUES ($1, $2)
		 ON CONFLICT (from_word) DO UPDATE SET to_word = EXCLUDED.to_word`,
		from, to,
	)
	return err
}

func (r *repo) DeleteWordRule(ctx context.Context, from string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM text_word_rules WHERE from_word = $1`,
		from,
	)
	return err
}
