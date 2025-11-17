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
	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	txt := strings.TrimSpace(userText)
	if txt == "" {
		return "", fmt.Errorf("empty userText")
	}

	// 1) История
	history, err := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] ⚠️ fitting history load error: %v", err)
	} else {
		log.Printf("[ai] ✔️ Fitting history loaded: %d records (GPT sees only trimmed history)", len(history))
	}

	// 2) Стиль
	stylePrompt, err := s.promptRepo.GetByBotID(ctx, botID)
	if err != nil || strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
		log.Printf("[ai] 🔹 stylePrompt: default used")
	} else {
		log.Printf("[ai] 🔹 stylePrompt loaded")
	}

	// 3) Жёсткая инструкция
	superPrompt := `У тебя есть промпт (стиль), история диалога и последнее сообщение. 
Ответь строго на последнее сообщение, учитывая историю и стиль.`

	// 4) Сборка массива сообщений
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

	// 5) Последний запрос юзера
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: txt,
	})

	log.Printf("[ai] 🧩 messages built for GPT: %d", len(messages))

	// 6) Отправляем в GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	if err != nil {
		log.Printf("[ai] ❌ GPT error: %v", err)
		return "", err
	}

	log.Printf("[ai] <<< OK reply received (%d chars)", len(reply))
	return reply, nil
}
