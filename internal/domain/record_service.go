package domain

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

type recordService struct {
	repo      ports.RecordRepo
	s3service ports.S3Service
}

func NewRecordService(repo ports.RecordRepo, s3 ports.S3Service) ports.RecordService {
	return &recordService{repo: repo, s3service: s3}
}

func (s *recordService) AddText(ctx context.Context, botID string, telegramID int64, role, text string) (int64, error) {
	id, err := s.repo.CreateText(ctx, botID, telegramID, role, text)
	if err != nil {
		return 0, err
	}

	// в фоне пересчитываем историю — не блокируем обработку
	go func() {
		if err := s.RecalcHistoryState(context.Background(), botID, telegramID); err != nil {
			log.Printf("[history] recalc fail: %v", err)
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
		return 0, "", err
	}

	id, err := s.repo.CreateImage(ctx, botID, telegramID, role, publicURL)
	if err != nil {
		return 0, "", err
	}

	go func() {
		if err := s.RecalcHistoryState(context.Background(), botID, telegramID); err != nil {
			log.Printf("[history] recalc fail: %v", err)
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

	limit := 90000 // потом вынести в конфиг
	totalTokens := 0
	lastN := 0

	enc, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Printf("tokenizer init fail: %v", err)
		return err
	}

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

	log.Printf("[history] bot=%s tg=%d lastN=%d tokens=%d", botID, telegramID, lastN, totalTokens)
	return nil
}

func countTokens(r ports.Record, enc *tiktoken.Tiktoken) int {
	switch {
	case r.Type == "text" && r.Text != nil:
		return len(enc.Encode(*r.Text, nil, nil))
	case r.Type == "image":
		return 60 // за картинку GPT обычно добавляет маленький overhead
	default:
		return 0
	}
}

func (s *recordService) GetFittingHistory(ctx context.Context, botID string, telegramID int64) ([]ports.Record, error) {
	lastN, _, err := s.repo.GetHistoryState(ctx, botID, telegramID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetLastNRecords(ctx, botID, telegramID, lastN)
}

// estimateTokens — можно прикинуть символы / 4, либо использовать реальный токенайзер
func estimateTokens(r ports.Record) int {
	if r.Text == nil || *r.Text == "" {
		return 0
	}
	return len(*r.Text) / 4 // грубая оценка: 1 токен ~ 4 символа
}
