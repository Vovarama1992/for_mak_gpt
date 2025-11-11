package ai

import (
	"context"
	"log"
	"os"
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
	log.Printf("[ai] >>> START botID=%s telegramID=%d userText=%q", botID, telegramID, userText)

	// сохраняем новое сообщение пользователя
	if txt := strings.TrimSpace(userText); txt != "" {
		if _, err := s.recordService.AddText(ctx, botID, telegramID, "user", txt); err != nil {
			log.Printf("[ai] save user text fail: %v", err)
		}
	}

	// достаём историю сообщений
	history, err := s.recordService.GetHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] history load fail: %v", err)
	} else {
		log.Printf("[ai] history fetched for botID=%s telegramID=%d -> %d records", botID, telegramID, len(history))
		for i, r := range history {
			if i >= 3 {
				log.Printf("[ai] ... (%d more records hidden)", len(history)-3)
				break
			}
			log.Printf("[ai] record[%d]: role=%s type=%s text=%q", i, r.Role, r.Type, safePtr(r.Text))
		}
	}

	// системный промпт из окружения
	systemPrompt := os.Getenv("OPENAI_SYSTEM_PROMPT")
	if systemPrompt == "" {
		systemPrompt = "Ты доброжелательная девушка-репетитор. Отвечай понятно, логично и с лёгким воодушевлением, помогая ученику разобраться в теме."
	}

	// формируем список сообщений
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: systemPrompt},
	}

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

	log.Printf("[ai] sending %d messages to GPT", len(messages))

	// получаем ответ от GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	if err != nil {
		log.Printf("[ai] GPT error: %v", err)
		return "", err
	}
	log.Printf("[ai] GPT reply: %q", reply)

	// сохраняем ответ
	if _, err := s.recordService.AddText(ctx, botID, telegramID, "tutor", reply); err != nil {
		log.Printf("[ai] save reply fail: %v", err)
	}

	log.Printf("[ai] <<< END botID=%s telegramID=%d", botID, telegramID)
	return reply, nil
}

func safePtr(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}
