package ai

import (
	"context"
	"log"
	"strings"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
}

func NewAiService(client *OpenAIClient, recordSvc ports.RecordService) *AiService {
	return &AiService{
		openaiClient:  client,
		recordService: recordSvc,
	}
}

func (s *AiService) GetReply(ctx context.Context, botID string, telegramID int64, userText string) (string, error) {
	// 0) Сохраняем вход пользователя (чтобы память была у КАЖДОГО бота отдельно)
	if txt := strings.TrimSpace(userText); txt != "" {
		if _, err := s.recordService.AddText(ctx, botID, telegramID, "user", txt); err != nil {
			log.Printf("[ai] save user text fail: %v", err)
		}
	}

	// 1) Тянем историю из БД
	history, err := s.recordService.GetHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] history load fail: %v", err)
	}

	// 2) Готовим сообщения для GPT
	var messages []openai.ChatCompletionMessage
	for _, r := range history {
		if r.Type != "text" || r.Text == nil || strings.TrimSpace(*r.Text) == "" {
			continue
		}
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: *r.Text,
		})
	}

	// 3) Запрос к GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	if err != nil {
		return "", err
	}

	// 4) Сохраняем ответ ассистента
	if _, err := s.recordService.AddText(ctx, botID, telegramID, "tutor", reply); err != nil {
		log.Printf("[ai] save reply fail: %v", err)
	}

	return reply, nil
}
