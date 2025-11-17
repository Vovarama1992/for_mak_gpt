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
	start := time.Now()

	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	txt := strings.TrimSpace(userText)
	if txt == "" {
		return "", fmt.Errorf("empty userText")
	}

	// 1) История
	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai] history entries: %d", len(history))

	// 2) Стиль
	stylePrompt, _ := s.promptRepo.GetByBotID(ctx, botID)
	if strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 3) Ищем последнюю картинку в истории
	var lastImageURL *string
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].ImageURL != nil && *history[i].ImageURL != "" {
			lastImageURL = history[i].ImageURL
			break
		}
	}

	// 4) Сборка сообщений
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: "Учитывай стиль, историю и отвечай только на последнее сообщение."},
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
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: strings.TrimSpace(*r.Text),
		})
	}

	// 5) Формируем итоговое сообщение — JSON если была картинка
	var content string
	if lastImageURL != nil {
		content = fmt.Sprintf(`[
			{"type": "text", "text": %q},
			{"type": "image_url", "image_url": {"url": %q}}
		]`, txt, *lastImageURL)
		log.Printf("[ai] using vision payload JSON with image: %s", *lastImageURL)
	} else {
		content = txt
		log.Printf("[ai] text-only payload")
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: content,
	})

	// 6) GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	log.Printf("[ai][%.1fs] GPT done, err=%v", time.Since(start).Seconds(), err)

	return reply, err
}
