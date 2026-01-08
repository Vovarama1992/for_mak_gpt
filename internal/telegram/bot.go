package telegram

import (
	"context"
	"log"
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
	// 0) СБРОС НАСТРОЕК
	// =====================================================
	if strings.Contains(textLower, "сброс") {
		log.Printf("[ui] RESET pressed tg=%d", tgID)

		if err := app.UserService.ResetUserSettings(ctx, botID, tgID); err != nil {
			log.Printf("[ui] reset error tg=%d err=%v", tgID, err)
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка сброса настроек."))
			return
		}

		status = "none"

		m := tgbotapi.NewMessage(chatID, "Настройки сброшены. Начнём заново.")
		m.ReplyMarkup = app.BuildMainKeyboard("none")
		bot.Send(m)
		return
	}

	// =====================================================
	// 1) ГЛОБАЛЬНЫЕ КНОПКИ
	// =====================================================

	if strings.Contains(textLower, "тариф") {
		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, app.BuildSubscriptionText(ctx, botID))
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	if strings.Contains(textLower, "минут") {
		menu := app.BuildMinutePackagesMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, "Выбери пакет минут:")
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	if strings.Contains(textLower, "помощ") {
		if app.adminBotUsername == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Поддержка временно недоступна."))
			return
		}

		url := "https://t.me/" + app.adminBotUsername + "?start=support"
		m := tgbotapi.NewMessage(chatID, "🆘 Написать в поддержку:")
		m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("✉️ Поддержка", url),
			),
		)
		bot.Send(m)
		return
	}

	// =====================================================
	// 2) НАЧАТЬ УРОК → ONBOARDING
	// =====================================================
	if strings.Contains(textLower, "начать") && status != "active" {
		cfg, _ := app.BotsService.Get(ctx, botID)

		if cfg != nil && cfg.WelcomeVideo != nil && *cfg.WelcomeVideo != "" {
			video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(*cfg.WelcomeVideo))
			bot.Send(video)
		}

		welcome := "Привет! Я — твой AI-репетитор 🤖"
		if cfg != nil && cfg.WelcomeText != nil {
			welcome = strings.TrimSpace(*cfg.WelcomeText)
		}
		bot.Send(tgbotapi.NewMessage(chatID, welcome))

		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, "Чтобы продолжить — выбери тариф:")
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	// =====================================================
	// 3) НЕТ ACTIVE → ВСЕГДА ТАРИФЫ
	// =====================================================
	if status != "active" {
		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, "⛔ Доступ закрыт. Выбери тариф:")
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	// =====================================================
	// 4) ACTIVE — КОНТЕНТ
	// =====================================================
	mainKB := app.BuildMainKeyboard("active")

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
		m := tgbotapi.NewMessage(chatID, "📎 Отправь текст, голос, фото или документ.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
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
