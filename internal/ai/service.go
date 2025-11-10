package ai

import (
	"context"
	"log"

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
	// 1. Получаем историю
	history, err := s.recordService.GetHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] history load fail: %v", err)
	}

	// 2. Преобразуем историю в формат GPT
	var messages []openai.ChatCompletionMessage
	for _, r := range history {
		if r.Type != "text" || r.Text == nil || *r.Text == "" {
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

	// 3. Добавляем новое сообщение пользователя
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: userText,
	})

	// 4. Отправляем в GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	if err != nil {
		return "", err
	}

	// 5. Сохраняем ответ в БД
	if _, err := s.recordService.AddText(ctx, botID, telegramID, "tutor", reply); err != nil {
		log.Printf("[ai] save reply fail: %v", err)
	}

	return reply, nil
}
