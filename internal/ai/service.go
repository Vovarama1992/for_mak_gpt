package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
	botsRepo      bots.Repo
	Notifier      error_notificator.Notificator
}

func NewAiService(
	client *OpenAIClient,
	recordSvc ports.RecordService,
	botsRepo bots.Repo,
	notifier error_notificator.Notificator,
) *AiService {
	return &AiService{
		openaiClient:  client,
		recordService: recordSvc,
		botsRepo:      botsRepo,
		Notifier:      notifier,
	}
}

// === точная диагностика ошибок OpenAI ===
func analyzeOpenAIError(err error) string {
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "status code: 401"):
		return "Неверный API-ключ OpenAI."
	case strings.Contains(msg, "status code: 404"):
		return "Модель не найдена — возможно, опечатка."
	case strings.Contains(msg, "status code: 429"):
		return "Превышен лимит OpenAI (rate limit)."
	case strings.Contains(msg, "status code: 400") &&
		strings.Contains(msg, "model"):
		return "Указана модель, которой не существует или она отключена."
	case strings.Contains(msg, "status code: 400"):
		return "Некорректный запрос к OpenAI."
	case strings.Contains(msg, "status code: 500"):
		return "Внутренняя ошибка OpenAI (HTTP 500)."
	}

	return "Неизвестная ошибка OpenAI: " + err.Error()
}

// === уведомление о проблемах с конфигом ===
func (s *AiService) notifyConfigError(ctx context.Context, botID string, err error) {
	details := fmt.Sprintf(
		"❗ Ошибка конфигурации бота\n\nБот: %s\nОшибка: %v\n\nЧто проверить:\n— model\n— style_prompt\n— voice_id\n— формат строки",
		botID, err,
	)
	s.Notifier.Notify(ctx, botID, err, details)
}

// === уведомление о GPT ошибках (с диагностикой) ===
func (s *AiService) notifyGptError(ctx context.Context, botID, model string, err error) {
	diagnosis := analyzeOpenAIError(err)

	details := fmt.Sprintf(
		"❗ Ошибка GPT\n\nБот: %s\nМодель: %s\nОшибка: %v\n\nДиагностика:\n%s",
		botID, model, err, diagnosis,
	)

	s.Notifier.Notify(ctx, botID, err, details)
}

// === основной метод ===
func (s *AiService) GetReply(
	ctx context.Context,
	botID string,
	telegramID int64,
	userText string,
	imageURL *string,
) (string, error) {

	start := time.Now()
	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	// 1. Конфиг бота
	cfg, err := s.botsRepo.Get(ctx, botID)
	if err != nil {
		s.notifyConfigError(ctx, botID, err)
		return "", err
	}

	stylePrompt := strings.TrimSpace(cfg.StylePrompt)
	if stylePrompt == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 2. История
	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai] history entries: %d", len(history))

	superPrompt := `У тебя есть промпт (стиль), история диалога и последнее сообщение.
Ответь строго на последнее сообщение, учитывая историю и стиль.`

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Промпт: " + stylePrompt},
	}

	// 3. История
	for _, r := range history {
		if r.Text == nil {
			continue
		}
		t := strings.TrimSpace(*r.Text)
		if t == "" {
			continue
		}
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: t,
		})
	}

	// 4. Последнее сообщение
	if imageURL == nil {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: userText,
		})
	} else {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: "user",
			MultiContent: []openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: userText},
				{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: *imageURL}},
			},
		})
	}

	// 5. GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages, cfg.Model)
	log.Printf("[ai][%.1fs] GPT done, err=%v", time.Since(start).Seconds(), err)

	if err != nil {
		s.notifyGptError(ctx, botID, cfg.Model, err)
		return "", err
	}

	return reply, nil
}
