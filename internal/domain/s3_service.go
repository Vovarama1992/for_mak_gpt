package domain

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type s3Service struct {
	client ports.S3Client
}

func NewS3Service(client ports.S3Client) ports.S3Service {
	return &s3Service{client: client}
}

// ObjectKey — экспортируемый, чтобы реализовывать интерфейс
func (s *s3Service) ObjectKey(telegramID int64, filename string) string {
	date := time.Now().Format("2006-01-02")
	clean := filepath.Base(filename)
	return fmt.Sprintf("%d/%s/%s", telegramID, date, clean)
}

// SaveImage — теперь с botID
func (s *s3Service) SaveImage(ctx context.Context, botID string, telegramID int64, file multipart.File, filename, contentType string) (string, error) {
	if botID == "" {
		return "", fmt.Errorf("botID required")
	}
	key := s.ObjectKey(telegramID, filename)
	return s.client.PutObject(ctx, botID, key, file, -1, contentType)
}
