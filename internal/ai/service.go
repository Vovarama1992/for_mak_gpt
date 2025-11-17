package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
	promptRepo    ports.PromptRepo
}

func NewAiService(client *OpenAIClient, recordSvc ports.RecordService, promptRepo ports.PromptRepo) *AiService {
	return &AiService{
		openaiClient:  client,
		recordService: recordSvc,
		promptRepo:    promptRepo,
	}
}

func (s *AiService) GetReply(ctx context.Context, botID string, telegramID int64, userText string) (string, error) {
	log.Printf("[ai] >>> START botID=%s telegramID=%d", botID, telegramID)

	txt := strings.TrimSpace(userText)
	if txt == "" {
		return "", fmt.Errorf("empty userText")
	}

	// 1) Берём готовую «вмещающуюся» историю
	history, err := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] history load fail: %v", err)
	}

	// 2) Системный промпт
	stylePrompt, err := s.promptRepo.GetByBotID(ctx, botID)
	if err != nil || strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 3) Жёсткая инструкция
	superPrompt := `У тебя есть промпт (стиль), история диалога и последнее сообщение. 
Ответь строго на последнее сообщение, учитывая историю и стиль. `

	// === Формируем запрос ===
	msg := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Промпт: " + stylePrompt},
	}

	for _, r := range history {
		if r.Text == nil || strings.TrimSpace(*r.Text) == "" {
			continue
		}
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		msg = append(msg, openai.ChatCompletionMessage{
			Role:    role,
			Content: strings.TrimSpace(*r.Text),
		})
	}

	// 4) Последний запрос
	msg = append(msg, openai.ChatCompletionMessage{
		Role:    "user",
		Content: txt,
	})

	// 5) → GPT
	reply, err := s.openaiClient.GetCompletion(ctx, msg)
	if err != nil {
		return "", err
	}

	return reply, nil
}
