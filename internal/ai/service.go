package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/classes"
	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
	botsRepo      bots.Repo
	classService  classes.ClassService
	Notifier      error_notificator.Notificator
}

func NewAiService(
	client *OpenAIClient,
	recordSvc ports.RecordService,
	botsRepo bots.Repo,
	classSvc classes.ClassService,
	notifier error_notificator.Notificator,
) *AiService {
	return &AiService{
		openaiClient:  client,
		recordService: recordSvc,
		botsRepo:      botsRepo,
		classService:  classSvc,
		Notifier:      notifier,
	}
}

// диагностика ошибок GPT
func analyzeOpenAIError(err error) string {
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "status code: 401"):
		return "Неверный API-ключ OpenAI."
	case strings.Contains(msg, "status code: 404"):
		return "Модель не найдена."
	case strings.Contains(msg, "status code: 429"):
		return "Превышен лимит OpenAI."
	case strings.Contains(msg, "status code: 400") && strings.Contains(msg, "model"):
		return "Неверно указана модель."
	case strings.Contains(msg, "status code: 400"):
		return "Некорректный запрос к OpenAI."
	case strings.Contains(msg, "status code: 500"):
		return "Внутренняя ошибка OpenAI."
	}
	return "Неизвестная ошибка OpenAI: " + err.Error()
}

// уведомления
func (s *AiService) notifyConfigError(ctx context.Context, botID string, err error) {
	s.Notifier.Notify(ctx, botID, err,
		fmt.Sprintf("Ошибка конфигурации бота %s: %v", botID, err))
}

func (s *AiService) notifyGptError(ctx context.Context, botID, model string, err error) {
	diag := analyzeOpenAIError(err)
	s.Notifier.Notify(ctx, botID, err,
		fmt.Sprintf("Ошибка GPT\nБот: %s\nМодель: %s\n%v\n\n%s",
			botID, model, err, diag))
}

// === главный метод ===
func (s *AiService) GetReply(
	ctx context.Context,
	botID string,
	telegramID int64,
	branch string, // может быть пустым
	userText string,
	imageURL *string,
) (string, error) {

	if branch == "" {
		branch = "text"
	}

	start := time.Now()
	log.Printf("[ai] >>> START bot=%s tg=%d branch=%s", botID, telegramID, branch)

	// 1) конфиг бота
	cfg, err := s.botsRepo.Get(ctx, botID)
	if err != nil {
		s.notifyConfigError(ctx, botID, err)
		return "", err
	}

	// 2) стилевой промпт по ветке
	var stylePrompt string
	switch branch {
	case "voice":
		stylePrompt = cfg.VoiceStylePrompt
	default:
		stylePrompt = cfg.TextStylePrompt
	}

	stylePrompt = strings.TrimSpace(stylePrompt)
	if stylePrompt == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 3) классовый промпт
	finalClassPrompt := ""

	uc, err := s.classService.GetUserClass(ctx, botID, telegramID)
	if err == nil && uc != nil {
		p, err := s.classService.GetPromptByClassID(ctx, uc.ClassID)
		if err == nil && p != nil {
			finalClassPrompt = strings.TrimSpace(p.Prompt)
		}
	}

	// 4) full style prompt
	fullStyle := stylePrompt
	if finalClassPrompt != "" {
		fullStyle += "\n\n" + finalClassPrompt
	}

	// 5) координационный супер-промпт
	superPrompt := `Это координационный промпт. 
У тебя есть стилевой промпт и промпт класса ученика. 
История — только контекст. 
Отвечай строго на последнее сообщение, ориентируясь на историю и промпты.`

	// 6) история
	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai] history entries: %d", len(history))

	// 7) формируем messages
	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Стилевой промпт: " + fullStyle},
	}

	for _, r := range history {
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}

		if r.Text != nil {
			txt := strings.TrimSpace(*r.Text)
			if txt != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    role,
					Content: txt,
				})
			}
		}

		if r.ImageURL != nil {
			url := strings.TrimSpace(*r.ImageURL)
			if url != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role: role,
					MultiContent: []openai.ChatMessagePart{
						{Type: openai.ChatMessagePartTypeText, Text: "(Ранее прислано изображение)"},
						{Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{URL: url}},
					},
				})
			}
		}
	}

	// 8) последнее сообщение
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
				{Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{URL: *imageURL}},
			},
		})
	}

	// 9) GPT
	reply, err := s.openaiClient.GetCompletion(ctx, messages, cfg.Model)
	log.Printf("[ai][%.1fs] GPT done err=%v", time.Since(start).Seconds(), err)

	if err != nil {
		s.notifyGptError(ctx, botID, cfg.Model, err)
		return "", err
	}

	return reply, nil
}
