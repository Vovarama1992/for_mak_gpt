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

func (s *AiService) GetReply(
	ctx context.Context,
	botID string,
	telegramID int64,
	userText string,
	imageURL *string, // nil если нет фото
) (string, error) {

	start := time.Now()
	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	// 1. Стиль
	stylePrompt, _ := s.promptRepo.GetByBotID(ctx, botID)
	if strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 2. История
	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai] history entries: %d", len(history))

	// 3. Системное указание
	superPrompt := `У тебя есть промпт (стиль), история диалога и последнее сообщение.
Ответь строго на последнее сообщение, учитывая историю и стиль.`

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Промпт: " + stylePrompt},
	}

	// 4. История
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

	// 5. Последнее сообщение
	if imageURL != nil {
		log.Printf("[ai] vision payload, image=%s", *imageURL)
		userText = fmt.Sprintf("%s\nВот изображение для анализа: %s", userText, *imageURL)
	} else {
		log.Printf("[ai] text-only payload")
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: userText,
	})

	// 6. GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	log.Printf("[ai][%.1fs] GPT done, err=%v", time.Since(start).Seconds(), err)

	return reply, err
}
