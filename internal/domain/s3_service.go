package domain

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	error_notificator "github.com/Vovarama1992/make_ziper/internal/notificator"
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

// путь для пользовательской истории (фото/pdf/etc)
func (s *s3Service) ObjectKey(telegramID int64, filename string) string {
	date := time.Now().Format("2006-01-02")
	clean := filepath.Base(filename)
	return fmt.Sprintf("%d/%s/%s", telegramID, date, clean)
}

// SaveImage — для пользовательских сообщений
func (s *s3Service) SaveImage(
	ctx context.Context,
	botID string,
	telegramID int64,
	file io.Reader,
	filename string,
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

// ======================================================
// SaveBotAsset — для ассетов бота (welcome video, картинки)
// путь: bots/{botID}/assets/<filename>
// ======================================================
func (s *s3Service) SaveBotAsset(
	ctx context.Context,
	botID string,
	file io.Reader,
	filename string,
) (string, error) {

	if botID == "" {
		err := fmt.Errorf("botID required")
		s.notifier.Notify(ctx, "unknown", err, "SaveBotAsset: botID пустой")
		return "", err
	}

	clean := filepath.Base(filename)
	key := fmt.Sprintf("bots/%s/assets/%s", botID, clean)

	url, err := s.client.PutObject(ctx, botID, key, file, -1, "video/mp4")
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка загрузки ассета: botID=%s key=%s", botID, key))
		return "", err
	}

	return url, nil
}
