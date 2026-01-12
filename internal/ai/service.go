package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/classes"
	notificator "github.com/Vovarama1992/make_ziper/internal/notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	openai "github.com/sashabaranov/go-openai"
)

type AiService struct {
	openaiClient  *OpenAIClient
	recordService ports.RecordService
	botsRepo      bots.Repo
	classService  classes.ClassService
	Notifier      notificator.Notificator
}

func NewAiService(
	client *OpenAIClient,
	recordSvc ports.RecordService,
	botsRepo bots.Repo,
	classSvc classes.ClassService,
	notifier notificator.Notificator,
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

	case "image", "photo":
		stylePrompt = cfg.PhotoStylePrompt

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
		p, err := s.classService.GetPromptByClassID(ctx, botID, uc.ClassID)
		if err == nil && p != nil {
			finalClassPrompt = strings.TrimSpace(p.Prompt)
		}
	}

	// 4) full style prompt
	fullStyle := stylePrompt
	if finalClassPrompt != "" {
		fullStyle += "\n\n" + finalClassPrompt
	}

	fullStyle += `
Ограничение: ответ не должен превышать 3000 символов.
Пиши кратко, структурировано. Если документ большой — дай краткое summary, главное — смысл.`
	// 5) координационный супер-промпт
	superPrompt := `Это координационный промпт.
Последний пользовательский запрос может состоять из одного сообщения ИЛИ из набора сообщений: текста и файлов (изображений, документов).

Если в последних нескольких сообщениях истории есть файлы, относящиеся к текущему запросу, считай их частью последнего пользовательского ввода.
Если в последних сообщениях файлов нет, то последним пользовательским вводом является ровно одно последнее сообщение.

Файлы в истории представлены в виде ссылок (URL).
Не анализируй и не учитывай файлы по ссылкам из истории автоматически.
Считай, что файлы из истории являются неактивным контекстом, если пользователь явно не указывает, что нужно использовать конкретный файл или изображение.

Используй ТОЛЬКО те файлы, которые явно относятся к текущему запросу.
Если нужно проанализировать файл из истории, пользователь должен прямо на него сослаться в тексте.

Не проси повторно файлы или изображения, если они уже присутствуют в истории и пользователь явно указал на них.
Отвечай на запрос, используя текст последнего сообщения и только связанные с ним файлы.
`

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
	if imageURL != nil {
		if strings.TrimSpace(userText) != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{Type: openai.ChatMessagePartTypeText, Text: userText},
					{
						Type:     openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{URL: *imageURL},
					},
				},
			})
		} else {
			messages = append(messages, openai.ChatCompletionMessage{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type:     openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{URL: *imageURL},
					},
				},
			})
		}
	} else {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: userText,
		})
	}

	// 9) GPT

	ctxGPT, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	reply, err := s.openaiClient.GetCompletion(ctxGPT, messages, cfg.Model)
	log.Printf("[ai][%.1fs] GPT done err=%v", time.Since(start).Seconds(), err)

	if err != nil {
		s.notifyGptError(ctx, botID, cfg.Model, err)
		return "", err
	}

	return reply, nil
}

func (s *AiService) GetReplyWithDirectImage(
	ctx context.Context,
	botID string,
	telegramID int64,
	userText string,
	imageURL string,
) (string, error) {

	start := time.Now()
	log.Printf("[ai] >>> START DIRECT_IMAGE bot=%s tg=%d", botID, telegramID)

	cfg, err := s.botsRepo.Get(ctx, botID)
	if err != nil {
		s.notifyConfigError(ctx, botID, err)
		return "", err
	}

	stylePrompt := strings.TrimSpace(cfg.PhotoStylePrompt)
	if stylePrompt == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	fullStyle := stylePrompt + `
Ограничение: ответ не должен превышать 3000 символов.
Пиши кратко и по делу.`

	superPrompt := `Это координационный промпт для прямого анализа изображения.

Последний пользовательский ввод ВСЕГДА состоит из одного сообщения с изображением (image_url) и, возможно, сопроводительного текста.
Это изображение является единственным активным файлом для анализа.

Игнорируй все ссылки, изображения и документы из истории.
НЕ анализируй и НЕ учитывай файлы по URL из предыдущих сообщений.
История используется только как текстовый контекст.

Отвечай, опираясь исключительно на текущее изображение и текст последнего сообщения.
Не пытайся искать или использовать другие изображения из истории.
Не проси повторно изображение.`

	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Стилевой промпт: " + fullStyle},
	}

	for _, r := range history {
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		if r.Text != nil && strings.TrimSpace(*r.Text) != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: *r.Text,
			})
		}
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role: "user",
		MultiContent: []openai.ChatMessagePart{
			{Type: openai.ChatMessagePartTypeText, Text: userText},
			{
				Type:     openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
			},
		},
	})

	ctxGPT, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	reply, err := s.openaiClient.GetCompletion(ctxGPT, messages, cfg.Model)
	log.Printf("[ai][%.1fs] DIRECT_IMAGE done err=%v", time.Since(start).Seconds(), err)

	if err != nil {
		s.notifyGptError(ctx, botID, cfg.Model, err)
		return "", err
	}

	return reply, nil
}

func (s *AiService) GetReplyPDFOptimized(
	ctx context.Context,
	botID string,
	telegramID int64,
	userText string,
	maxImages int, // например 1 или 2
) (string, error) {

	start := time.Now()
	log.Printf("[ai] >>> START PDF_OPT bot=%s tg=%d", botID, telegramID)

	cfg, err := s.botsRepo.Get(ctx, botID)
	if err != nil {
		s.notifyConfigError(ctx, botID, err)
		return "", err
	}

	stylePrompt := strings.TrimSpace(cfg.PhotoStylePrompt)
	if stylePrompt == "" {
		stylePrompt = "Ты дружелюбный логичный ассистент."
	}

	fullStyle := stylePrompt + `
Ограничение: ответ не должен превышать 3000 символов.
Если документ большой — дай summary.`

	superPrompt := `Это координационный промпт.
Документ представлен изображениями в истории.
Используй только последние релевантные страницы.`

	history, _ := s.recordService.GetFittingHistory(ctx, botID, telegramID)

	// --- собираем последние N image_url ---
	imageURLs := make([]string, 0, maxImages)
	for i := len(history) - 1; i >= 0 && len(imageURLs) < maxImages; i-- {
		if history[i].ImageURL != nil {
			imageURLs = append(imageURLs, *history[i].ImageURL)
		}
	}

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: superPrompt},
		{Role: "system", Content: "Стилевой промпт: " + fullStyle},
	}

	// текстовую историю добавляем всю
	for _, r := range history {
		role := "user"
		if r.Role == "tutor" {
			role = "assistant"
		}
		if r.Text != nil && strings.TrimSpace(*r.Text) != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: *r.Text,
			})
		}
	}

	// добавляем ТОЛЬКО последние N картинок
	for _, url := range imageURLs {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: "user",
			MultiContent: []openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: "(страница документа)"},
				{
					Type:     openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{URL: url},
				},
			},
		})
	}

	// финальный текстовый запрос
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: userText,
	})

	ctxGPT, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	reply, err := s.openaiClient.GetCompletion(ctxGPT, messages, cfg.Model)
	log.Printf("[ai][%.1fs] PDF_OPT done err=%v", time.Since(start).Seconds(), err)

	if err != nil {
		s.notifyGptError(ctx, botID, cfg.Model, err)
		return "", err
	}

	return reply, nil
}
