package domain

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

type recordService struct {
	repo      ports.RecordRepo
	s3service ports.S3Service
	notifier  error_notificator.Notificator
}

func NewRecordService(repo ports.RecordRepo, s3 ports.S3Service, n error_notificator.Notificator) ports.RecordService {
	return &recordService{
		repo:      repo,
		s3service: s3,
		notifier:  n,
	}
}

func (s *recordService) AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error) {
	id, err := s.repo.CreateText(ctx, botID, telegramID, role, text)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка записи текста в history: tg=%d", telegramID))
		return 0, err
	}

	go func() {
		if err := s.RecalcHistoryState(context.Background(), botID, telegramID); err != nil {
			s.notifier.Notify(context.Background(), botID, err,
				fmt.Sprintf("Ошибка перерасчёта истории: tg=%d", telegramID))
		}
	}()

	return id, nil
}

func (s *recordService) AddImage(
	ctx context.Context, botID string, telegramID int64, role string,
	file io.Reader, filename, contentType string,
) (int64, string, error) {

	publicURL, err := s.s3service.SaveImage(ctx, botID, telegramID, file, filename, contentType)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка загрузки фото в S3: tg=%d filename=%s", telegramID, filename))
		return 0, "", err
	}

	id, err := s.repo.CreateImage(ctx, botID, telegramID, role, publicURL)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка записи image record: tg=%d url=%s", telegramID, publicURL))
		return 0, "", err
	}

	go func() {
		if err := s.RecalcHistoryState(context.Background(), botID, telegramID); err != nil {
			s.notifier.Notify(context.Background(), botID, err,
				fmt.Sprintf("Ошибка перерасчёта истории после фото: tg=%d", telegramID))
		}
	}()

	return id, publicURL, nil
}

func (s *recordService) GetHistory(ctx context.Context, botID string, telegramID int64) ([]ports.Record, error) {
	return s.repo.GetHistory(ctx, botID, telegramID)
}

func (s *recordService) ListUsers(ctx context.Context) ([]ports.UserBots, error) {
	return s.repo.ListUsers(ctx)
}

func (s *recordService) RecalcHistoryState(ctx context.Context, botID string, telegramID int64) error {
	records, err := s.repo.GetHistory(ctx, botID, telegramID)
	if err != nil {
		return err
	}

	enc, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Printf("tokenizer init fail: %v", err)
		return err
	}

	limit := 90000
	totalTokens := 0
	lastN := 0

	for i := len(records) - 1; i >= 0; i-- {
		tokens := countTokens(records[i], enc)
		if totalTokens+tokens > limit {
			break
		}
		totalTokens += tokens
		lastN++
	}

	if err := s.repo.UpsertHistoryState(ctx, botID, telegramID, lastN, totalTokens); err != nil {
		return fmt.Errorf("update history state fail: %w", err)
	}

	return nil
}

func countTokens(r ports.Record, enc *tiktoken.Tiktoken) int {
	switch {
	case r.Type == "text" && r.Text != nil:
		return len(enc.Encode(*r.Text, nil, nil))
	case r.Type == "image":
		return 60
	}
	return 0
}

func (s *recordService) GetFittingHistory(
	ctx context.Context,
	botID string,
	telegramID int64,
) ([]ports.Record, error) {

	lastN, _, err := s.repo.GetHistoryState(ctx, botID, telegramID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetLastNRecords(ctx, botID, telegramID, lastN)
}
