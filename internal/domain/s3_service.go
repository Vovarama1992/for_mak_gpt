package domain

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type s3Service struct {
	client   ports.S3Client
	notifier error_notificator.Notificator
}

func NewS3Service(client ports.S3Client, n error_notificator.Notificator) ports.S3Service {
	return &s3Service{
		client:   client,
		notifier: n,
	}
}

// ObjectKey — путь в бакете
func (s *s3Service) ObjectKey(telegramID int64, filename string) string {
	date := time.Now().Format("2006-01-02")
	clean := filepath.Base(filename)
	return fmt.Sprintf("%d/%s/%s", telegramID, date, clean)
}

// SaveImage — принимает io.Reader
func (s *s3Service) SaveImage(
	ctx context.Context,
	botID string,
	telegramID int64,
	file io.Reader,
	filename,
	contentType string,
) (string, error) {

	if botID == "" {
		err := fmt.Errorf("botID required")
		s.notifier.Notify(ctx, "unknown", err, "S3Service: botID пустой")
		return "", err
	}

	key := s.ObjectKey(telegramID, filename)

	url, err := s.client.PutObject(ctx, botID, key, file, -1, contentType)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка загрузки в S3: tg=%d key=%s", telegramID, key))
		return "", err
	}

	return url, nil
}
