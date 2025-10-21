package infra

import (
	"context"
	"database/sql"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type recordRepo struct {
	db *sql.DB
}

func NewRecordRepo(db *sql.DB) ports.RecordRepo {
	return &recordRepo{db: db}
}

func (r *recordRepo) CreateText(ctx context.Context, telegramID int64, role, text string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO records (telegram_id, role, record_type, text_content, created_at)
		VALUES ($1, $2, 'text', $3, $4)
		RETURNING id
	`, telegramID, role, text, time.Now()).Scan(&id)
	return id, err
}

func (r *recordRepo) CreateImage(ctx context.Context, telegramID int64, role, imageURL string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO records (telegram_id, role, record_type, image_url, created_at)
		VALUES ($1, $2, 'image', $3, $4)
		RETURNING id
	`, telegramID, role, imageURL, time.Now()).Scan(&id)
	return id, err
}

func (r *recordRepo) GetHistory(ctx context.Context, telegramID int64) ([]ports.Record, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, telegram_id, user_ref, role, record_type, text_content, image_url, created_at
		FROM records
		WHERE telegram_id = $1
		ORDER BY created_at ASC
	`, telegramID)
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

	// если история меньше 60k символов — возвращаем всё
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

	// если больше — оставляем последние в пределах лимита
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
