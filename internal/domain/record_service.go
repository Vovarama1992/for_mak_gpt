package domain

import (
	"context"
	"fmt"
	"log"

	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

type recordService struct {
	repo     ports.RecordRepo
	notifier error_notificator.Notificator
}

func NewRecordService(repo ports.RecordRepo, n error_notificator.Notificator) ports.RecordService {
	return &recordService{
		repo:     repo,
		notifier: n,
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
		_ = s.RecalcHistoryState(context.Background(), botID, telegramID)
	}()

	return id, nil
}

func (s *recordService) AddImage(
	ctx context.Context,
	botID string,
	telegramID int64,
	role string,
	imageURL string,
) (int64, error) {

	id, err := s.repo.CreateImage(ctx, botID, telegramID, role, imageURL)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка записи image record: tg=%d url=%s", telegramID, imageURL))
		return 0, err
	}

	go func() {
		_ = s.RecalcHistoryState(context.Background(), botID, telegramID)
	}()

	return id, nil
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

	return s.repo.UpsertHistoryState(ctx, botID, telegramID, lastN, totalTokens)
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

func (s *recordService) DeleteAll(ctx context.Context) error {
	return s.repo.DeleteAll(ctx)
}

func (s *recordService) DeleteUserHistory(ctx context.Context, botID string, telegramID int64) error {
	err := s.repo.DeleteByUser(ctx, botID, telegramID)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка очистки истории: tg=%d", telegramID))
		return err
	}

	// сбрасываем state
	return s.repo.UpsertHistoryState(ctx, botID, telegramID, 0, 0)
}
