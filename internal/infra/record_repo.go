package infra

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type recordRepo struct {
	db *sql.DB
}

func NewRecordRepo(db *sql.DB) ports.RecordRepo {
	return &recordRepo{db: db}
}

func (r *recordRepo) CreateText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO records (bot_id, telegram_id, role, record_type, text_content, created_at)
		VALUES ($1, $2, $3, 'text', $4, $5)
		RETURNING id
	`, botID, telegramID, role, text, time.Now()).Scan(&id)
	return id, err
}

func (r *recordRepo) CreateImage(ctx context.Context, botID string, telegramID int64, role, imageURL string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO records (bot_id, telegram_id, role, record_type, image_url, created_at)
		VALUES ($1, $2, $3, 'image', $4, $5)
		RETURNING id
	`, botID, telegramID, role, imageURL, time.Now()).Scan(&id)
	return id, err
}

func (r *recordRepo) GetHistory(ctx context.Context, botID string, telegramID int64) ([]ports.Record, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, telegram_id, bot_id, user_ref, role, record_type, text_content, image_url, created_at
		FROM records
		WHERE telegram_id = $1 AND bot_id = $2
		ORDER BY created_at ASC
	`, telegramID, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ports.Record
	for rows.Next() {
		var rec ports.Record
		if err := rows.Scan(
			&rec.ID,
			&rec.TelegramID,
			&rec.BotID,
			&rec.UserRef,
			&rec.Role,
			&rec.Type,
			&rec.Text,
			&rec.ImageURL,
			&rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// обрезаем историю по лимиту символов
	const maxChars = 60000
	total := 0
	for _, rec := range records {
		if rec.Text != nil {
			total += len(*rec.Text)
		}
	}
	if total <= maxChars {
		return records, nil
	}

	total = 0
	start := len(records)
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Text != nil {
			total += len(*records[i].Text)
		}
		if total > maxChars {
			start = i + 1
			break
		}
	}
	if start > len(records) {
		start = len(records)
	}
	return records[start:], nil
}

func (r *recordRepo) ListUsers(ctx context.Context) ([]ports.UserBots, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT telegram_id, array_agg(DISTINCT bot_id) AS bots
		FROM records
		GROUP BY telegram_id
		ORDER BY telegram_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.UserBots
	for rows.Next() {
		var u ports.UserBots
		if err := rows.Scan(&u.TelegramID, pq.Array(&u.Bots)); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *recordRepo) UpsertHistoryState(ctx context.Context, botID string, telegramID int64, lastN, totalTokens int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO history_state (bot_id, telegram_id, last_n_records, total_tokens, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (bot_id, telegram_id)
		DO UPDATE SET last_n_records = $3, total_tokens = $4, updated_at = NOW()
	`, botID, telegramID, lastN, totalTokens)
	return err
}

func (r *recordRepo) GetHistoryState(ctx context.Context, botID string, telegramID int64) (int, int, error) {
	var lastN, totalTokens int
	err := r.db.QueryRowContext(ctx, `
		SELECT last_n_records, total_tokens
		FROM history_state
		WHERE bot_id = $1 AND telegram_id = $2
	`, botID, telegramID).Scan(&lastN, &totalTokens)
	return lastN, totalTokens, err
}

func (r *recordRepo) GetLastNRecords(ctx context.Context, botID string, telegramID int64, n int) ([]ports.Record, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, telegram_id, bot_id, user_ref, role, record_type, text_content, image_url, created_at
		FROM records
		WHERE telegram_id = $1 AND bot_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, telegramID, botID, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ports.Record
	for rows.Next() {
		var rec ports.Record
		if err := rows.Scan(
			&rec.ID,
			&rec.TelegramID,
			&rec.BotID,
			&rec.UserRef,
			&rec.Role,
			&rec.Type,
			&rec.Text,
			&rec.ImageURL,
			&rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// хронологический порядок
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	return records, nil
}
