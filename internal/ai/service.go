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
	log.Printf("[ai] >>> START botID=%s telegramID=%d userText=%q", botID, telegramID, userText)

	txt := strings.TrimSpace(userText)
	if txt != "" {
		if _, err := s.recordService.AddText(ctx, botID, telegramID, "user", txt); err != nil {
			log.Printf("[ai] save user text fail: %v", err)
		}
	}

	// грузим старую историю — но тримаем
	history, err := s.recordService.GetHistory(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[ai] history load fail: %v", err)
		history = nil
	}

	// системный промпт
	systemPrompt, err := s.promptRepo.GetByBotID(ctx, botID)
	if err != nil || strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = "Ты доброжелательная девушка-репетитор. Отвечай понятно, логично и с лёгким воодушевлением."
	}

	// ключевое правило – ВСЕГДА отвечать на последний вопрос, без перенаправлений
	overridePrompt := "Важно: отвечай строго на ПОСЛЕДНИЙ вопрос пользователя, независимо от истории. Не предлагай меню, не перенаправляй, не упоминай 'профиль'. Просто дай лучший возможный ответ по теме вопроса."

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "system", Content: overridePrompt},
	}

	// берём до 10 последних текстовых сообщений
	limit := 10
	if len(history) > limit {
		history = history[len(history)-limit:]
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

	// последний вопрос всегда в конце (самый важный)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: txt,
	})

	log.Printf("[ai] sending %d messages to GPT", len(messages))

	reply, err := s.openaiClient.GetCompletion(ctx, messages)
	if err != nil {
		log.Printf("[ai] GPT error: %v", err)
		return "", err
	}

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
