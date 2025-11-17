package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
	startTotal := time.Now()
	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	txt := strings.TrimSpace(userText)
	if txt == "" {
		return "", fmt.Errorf("empty userText")
	}

	// 1. История
	history, err := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai][t=%.1fs] history loaded: %d records (err=%v)", time.Since(startTotal).Seconds(), len(history), err)

	// 2. Стиль
	stylePrompt, err := s.promptRepo.GetByBotID(ctx, botID)
	if err != nil || strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// Ищем последнюю картинку в истории
	var lastImageURL *string
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].ImageURL != nil && *history[i].ImageURL != "" {
			lastImageURL = history[i].ImageURL
			break
		}
	}

	// 3. Сборка GPT messages
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: "У тебя есть промпт (стиль), история диалога и последнее сообщение. Ответь только на последнее сообщение."},
		{Role: "system", Content: "Промпт: " + stylePrompt},
	}

	// История (только текст)
	for _, r := range history {
		if r.Text == nil || strings.TrimSpace(*r.Text) == "" {
			continue
		}
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		messages = append(messages, openai.ChatCompletionMessage{Role: role, Content: strings.TrimSpace(*r.Text)})
	}

	// Последнее сообщение — JSON, если есть картинка
	var content string
	if lastImageURL != nil {
		content = fmt.Sprintf(`[
			{"type": "text", "text": %q},
			{"type": "image_url", "image_url": {"url": %q}}
		]`, txt, *lastImageURL)
	} else {
		content = txt
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: content,
	})

	// 4. GPT запрос
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	log.Printf("[ai][t=%.1fs] GPT responded, err=%v", time.Since(startTotal).Seconds(), err)

	if err != nil {
		return "", err
	}

	log.Printf("[ai] <<< DONE total=%.1fs reply chars=%d", time.Since(startTotal).Seconds(), len(reply))
	return reply, nil
}
