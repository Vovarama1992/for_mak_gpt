package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// runBotLoop — главный цикл получения апдейтов
func (app *BotApp) runBotLoop(botID string, bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	log.Printf("[bot_loop] started botID=%s username=@%s", botID, bot.Self.UserName)

	for update := range updates {

		var fromID int64
		switch {
		case update.Message != nil && update.Message.From != nil:
			fromID = update.Message.From.ID
		case update.CallbackQuery != nil && update.CallbackQuery.From != nil:
			fromID = update.CallbackQuery.From.ID
		}

		if fromID != 0 {
			log.Printf(
				"[bot_touch] botID=%s fromTG=%d updateID=%d",
				botID,
				fromID,
				update.UpdateID,
			)
		}

		ctx := context.Background()

		tgID := extractTelegramID(update)
		if tgID == 0 {
			continue
		}

		status, err := app.SubscriptionService.GetStatus(ctx, botID, tgID)
		if err != nil {
			log.Printf(
				"[bot_loop] getStatus fail botID=%s tgID=%d err=%v",
				botID,
				tgID,
				err,
			)
			continue
		}

		app.dispatchUpdate(ctx, botID, bot, tgID, status, update)
	}
}

func (app *BotApp) dispatchUpdate(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	tgID int64,
	status string,
	update tgbotapi.Update,
) {
	switch {
	case update.Message != nil:
		if botID == "perplexity" {
			app.handlePerplexity(ctx, bot, update.Message)
			return
		}
		app.handleMessage(ctx, botID, bot, update.Message, tgID, status)

	case update.CallbackQuery != nil:
		app.handleCallback(ctx, botID, bot, update.CallbackQuery, status)
	}
}

func (app *BotApp) handleMessage(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	status string,
) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)
	textLower := strings.ToLower(text)

	log.Printf("[handleMessage] tg=%d status=%s text=%q", tgID, status, text)

	// =====================================================
	// 0) КЛАВИАТУРА ВСЕГДА
	// =====================================================
	anchor := tgbotapi.NewMessage(chatID, " ")
	anchor.ReplyMarkup = app.BuildMainKeyboard(botID, status)
	bot.Send(anchor)

	// =====================================================
	// 1) СБРОС НАСТРОЕК
	// =====================================================
	if strings.Contains(textLower, "заново") {
		if err := app.UserService.ResetUserSettings(ctx, botID, tgID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сброса настроек."))
			return
		}

		bot.Send(tgbotapi.NewMessage(chatID, "Настройки сброшены."))

		app.ShowClassPicker(ctx, botID, bot, tgID, chatID)
		return
	}

	// =====================================================
	// 2) ГЛОБАЛЬНЫЕ КНОПКИ
	// =====================================================
	if strings.Contains(textLower, "ариф") {
		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, "💳 Выбери тариф:")
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}
	if textLower == "📦 остаток минут" {
		sub, _ := app.SubscriptionService.Get(ctx, botID, tgID)

		text := "🎧 У тебя осталось: 0.00 минут голосовых объяснений"
		if sub != nil {
			text = fmt.Sprintf(
				"🎧 У тебя осталось: %.2f голосовых объяснений\n\nПакеты минут:",
				sub.VoiceMinutes,
			)
		}

		menu := app.BuildMinutePackagesMenu(ctx, botID, tgID)

		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	if strings.Contains(textLower, "помощ") {
		if app.adminBotUsername == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Поддержка недоступна."))
			return
		}
		url := "https://t.me/" + app.adminBotUsername + "?start=support"
		m := tgbotapi.NewMessage(chatID, "🆘 Поддержка:")
		m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("✉️ Написать", url),
			),
		)
		bot.Send(m)
		return
	}

	// =====================================================
	// X) ОЧИСТКА ДИАЛОГА
	// =====================================================
	if strings.Contains(textLower, "очист") {
		if err := app.RecordService.DeleteUserHistory(ctx, botID, tgID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка очистки диалога."))
			return
		}

		m := tgbotapi.NewMessage(chatID, "🗑 Диалог очищен.")
		m.ReplyMarkup = app.BuildMainKeyboard(botID, "none")
		bot.Send(m)
		return
	}

	// =====================================================
	// 3) ПОДПИСКА ЕСТЬ → ОБЫЧНЫЙ ДИАЛОГ
	// =====================================================
	if status == "active" {
		mainKB := app.BuildMainKeyboard(botID, "active")

		switch {
		case msg.Voice != nil:
			app.handleVoice(ctx, botID, bot, msg, tgID, mainKB)
		case msg.Document != nil:
			if isPDF(msg.Document) {
				app.handlePDF(ctx, botID, bot, msg, tgID, mainKB)
			} else if isWord(msg.Document) {
				app.handleDoc(ctx, botID, bot, msg, tgID, mainKB)
			} else {
				app.handlePhoto(ctx, botID, bot, msg, tgID, mainKB)
			}
		case len(msg.Photo) > 0:
			app.handlePhoto(ctx, botID, bot, msg, tgID, mainKB)
		case text != "":
			app.handleText(ctx, botID, bot, msg, tgID, mainKB)
		default:
			m := tgbotapi.NewMessage(chatID, "Отправь текст, голос или файл.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
		}
		return
	}

	// =====================================================
	// 4) ПОДПИСКИ НЕТ
	// =====================================================
	trialUsed, err := app.TrialRepo.Exists(ctx, botID, tgID)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err, "Ошибка проверки trial")
		return
	}

	// --- 4.1 TRIAL УЖЕ БЫЛ → СРАЗУ ПЛАТНЫЕ ТАРИФЫ
	if trialUsed {
		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(
			chatID,
			"⛔ Пробный тариф уже использован.\nВыбери тариф:",
		)
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	// --- 4.2 TRIAL НЕ БЫЛ → ОНБОРДИНГ + ВЫБОР КЛАССА
	cfg, _ := app.BotsService.Get(ctx, botID)

	// приветственное видео
	if cfg != nil && cfg.WelcomeVideo != nil && *cfg.WelcomeVideo != "" {
		video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(*cfg.WelcomeVideo))
		video.ReplyMarkup = app.BuildMainKeyboard(botID, status)
		bot.Send(video)
	}

	// приветственный текст
	welcome := "Привет! Я — твой AI-репетитор 🤖"
	if cfg != nil && cfg.WelcomeText != nil {
		welcome = strings.TrimSpace(*cfg.WelcomeText)
	}

	msgOut := tgbotapi.NewMessage(chatID, welcome)
	msgOut.ReplyMarkup = app.BuildMainKeyboard(botID, status)
	bot.Send(msgOut)

	// активируем trial
	trial, err := app.TariffService.GetTrial(ctx, botID)
	if err != nil || trial == nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❗ Пробный тариф недоступен."))
		return
	}

	if err := app.SubscriptionService.ActivateTrial(
		ctx,
		botID,
		tgID,
		trial.Code,
	); err != nil {
		app.ErrorNotify.Notify(ctx, botID, err, "Ошибка активации trial")
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось активировать пробный тариф."))
		return
	}

	// выбор класса
	app.ShowClassPicker(ctx, botID, bot, tgID, chatID)

}

func (app *BotApp) handlePerplexity(
	ctx context.Context,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	sentThinking, _ := bot.Send(thinking)

	// ================= VOICE =================
	if msg.Voice != nil {
		fileID := msg.Voice.FileID

		file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое."))
			return
		}

		resp, err := http.Get(file.Link(bot.Token))
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки голосового."))
			return
		}
		defer resp.Body.Close()

		path := fmt.Sprintf("/tmp/%s.ogg", fileID)
		out, _ := os.Create(path)
		io.Copy(out, resp.Body)
		out.Close()
		defer os.Remove(path)

		text, err := app.SpeechService.Transcribe(ctx, "perplexity", path)
		if err != nil {
			bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос."))
			return
		}

		reply, err := app.AiService.GetPerplexityReply(ctx, text)
		if err != nil {
			bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка Perplexity."))
			return
		}

		outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
		if err := app.PerplexityTTS.Synthesize(ctx, reply, outVoice); err != nil {
			bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
			bot.Send(tgbotapi.NewMessage(chatID, reply))
			return
		}
		defer os.Remove(outVoice)

		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice)))
		return
	}

	// ================= TEXT =================
	if strings.TrimSpace(msg.Text) != "" {
		reply, err := app.AiService.GetPerplexityReply(ctx, msg.Text)
		if err != nil {
			bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка Perplexity."))
			return
		}

		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(tgbotapi.NewMessage(chatID, reply))
	}
}

func extractTelegramID(u tgbotapi.Update) int64 {
	switch {
	case u.Message != nil && u.Message.From != nil:
		return u.Message.From.ID
	case u.CallbackQuery != nil && u.CallbackQuery.From != nil:
		return u.CallbackQuery.From.ID
	default:
		return 0
	}
}

func (app *BotApp) checkVoiceAllowed(ctx context.Context, botID string, tgID int64) bool {
	sub, err := app.SubscriptionService.Get(ctx, botID, tgID)
	if err != nil || sub == nil {
		return false
	}
	if sub.Status != "active" {
		return false
	}
	return sub.VoiceMinutes > 0
}

func (app *BotApp) checkImageAllowed(ctx context.Context, botID string, tgID int64) bool {
	return true
}

func isPDF(doc *tgbotapi.Document) bool {
	name := strings.ToLower(doc.FileName)
	return strings.HasSuffix(name, ".pdf")
}

func isWord(doc *tgbotapi.Document) bool {
	name := strings.ToLower(doc.FileName)
	return strings.HasSuffix(name, ".doc") || strings.HasSuffix(name, ".docx")
}
