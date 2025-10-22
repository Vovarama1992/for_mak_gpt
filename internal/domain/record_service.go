package domain

import (
	"context"
	"mime/multipart"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type recordService struct {
	repo      ports.RecordRepo
	s3service ports.S3Service
}

func NewRecordService(repo ports.RecordRepo, s3 ports.S3Service) ports.RecordService {
	return &recordService{repo: repo, s3service: s3}
}

func (s *recordService) AddText(ctx context.Context, botID *string, telegramID int64, role, text string) (int64, error) {
	return s.repo.CreateText(ctx, botID, telegramID, role, text)
}

func (s *recordService) AddImage(ctx context.Context, botID *string, telegramID int64, role string, file multipart.File, filename string, contentType string) (int64, error) {
	publicURL, err := s.s3service.SaveImage(ctx, telegramID, file, filename, contentType)
	if err != nil {
		return 0, err
	}
	return s.repo.CreateImage(ctx, botID, telegramID, role, publicURL)
}

func (s *recordService) GetHistory(ctx context.Context, botID *string, telegramID int64) ([]ports.Record, error) {
	return s.repo.GetHistory(ctx, botID, telegramID)
}
