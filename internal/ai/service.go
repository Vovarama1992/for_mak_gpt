package ai

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
	botsRepo      bots.Repo
}

func NewAiService(
	client *OpenAIClient,
	recordSvc ports.RecordService,
	botsRepo bots.Repo,
) *AiService {
	return &AiService{
		openaiClient:  client,
		recordService: recordSvc,
		botsRepo:      botsRepo,
	}
}

func (s *AiService) GetReply(
	ctx context.Context,
	botID string,
	telegramID int64,
	userText string,
	imageURL *string,
) (string, error) {

	start := time.Now()
	log.Printf("[ai] >>> START bot=%s tg=%d", botID, telegramID)

	// 1. Берём конфиг бота (модель, стиль)
	cfg, err := s.botsRepo.Get(ctx, botID)
	if err != nil {
		log.Printf("[ai] bot config not found: %s", botID)
		return "", err
	}

	stylePrompt := strings.TrimSpace(cfg.StylePrompt)
	if stylePrompt == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	// 2. История
	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)
	log.Printf("[ai] history entries: %d", len(history))

	// 3. базовый системный промпт
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

	// 5. Последнее сообщение (текст или картинка)
	if imageURL == nil {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: userText,
		})
	} else {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: "user",
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: userText,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: *imageURL,
					},
				},
			},
		})
	}

	// 6. Вызов GPT с МОДЕЛЬЮ ИЗ БД
	reply, err := s.openaiClient.GetCompletion(ctx, messages, cfg.Model)
	log.Printf("[ai][%.1fs] GPT done, err=%v", time.Since(start).Seconds(), err)

	return reply, err
}
