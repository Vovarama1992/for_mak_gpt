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
	return records, rows.Err()
}
