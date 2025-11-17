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

	// 1) История
	t1 := time.Now()
	history, err := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai][t=%.1fs] history loaded: %d records (err=%v)",
		time.Since(t1).Seconds(), len(history), err)

	// 2) Стиль
	t2 := time.Now()
	stylePrompt, err := s.promptRepo.GetByBotID(ctx, botID)
	if err != nil || strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}
	log.Printf("[ai][t=%.1fs] stylePrompt resolved", time.Since(t2).Seconds())

	// 3) Сборка сообщений
	t3 := time.Now()
	superPrompt := `У тебя есть промпт (стиль), история диалога и последнее сообщение. 
	Ответь строго на последнее сообщение, учитывая историю и стиль.`

	messages := []openai.ChatCompletionMessage{
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
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: strings.TrimSpace(*r.Text),
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: txt,
	})

	log.Printf("[ai][t=%.1fs] messages built: %d msgs", time.Since(t3).Seconds(), len(messages))

	// 4) GPT запрос
	t4 := time.Now()
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	log.Printf("[ai][t=%.1fs] GPT responded, err=%v", time.Since(t4).Seconds(), err)

	if err != nil {
		return "", err
	}

	log.Printf("[ai] <<< DONE total=%.1fs reply chars=%d", time.Since(startTotal).Seconds(), len(reply))
	return reply, nil
}
