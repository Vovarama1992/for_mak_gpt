package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// runBotLoop — главный цикл получения апдейтов
func (app *BotApp) runBotLoop(botID string, bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	log.Printf("[bot_loop] started botID=%s username=@%s", botID, bot.Self.UserName)

	for update := range updates {
		ctx := context.Background()
		tgID := extractTelegramID(update)
		if tgID == 0 {
			continue
		}

		status, err := app.SubscriptionService.GetStatus(ctx, botID, tgID)
		if err != nil {
			log.Printf("[bot_loop] getStatus fail botID=%s tgID=%d err=%v", botID, tgID, err)
			continue
		}

		app.dispatchUpdate(ctx, botID, bot, tgID, status, update)
	}
}

func (app *BotApp) dispatchUpdate(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	tgID int64, status string, update tgbotapi.Update) {

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

	log.Printf("[sub-check] botID=%s tgID=%d → status=%s", botID, tgID, status)

	switch status {

	case "none":
		log.Printf("[none] botID=%s tgID=%d → show Start", botID, tgID)

		// Клавиатура "Старт"
		startKB := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("▶️ Старт"),
			),
		)
		startKB.ResizeKeyboard = true

		// Если нажали "Старт" → показать тарифы
		if msg.Text == "▶️ Старт" {
			menu := app.BuildSubscriptionMenu(ctx)
			text := app.BuildSubscriptionText()

			out := tgbotapi.NewMessage(chatID, text)
			out.ReplyMarkup = menu
			bot.Send(out)
			return
		}

		// Показываем кнопку старта
		welcome := tgbotapi.NewMessage(chatID,
			"Добро пожаловать! Нажми «Старт», чтобы выбрать тариф.")
		welcome.ReplyMarkup = startKB
		bot.Send(welcome)
		return

	case "pending":
		log.Printf("[pending] botID=%s tgID=%d → MsgPending", botID, tgID)
		bot.Send(tgbotapi.NewMessage(chatID, MsgPending))
		return

	case "expired":
		log.Printf("[expired] botID=%s tgID=%d → ask renew", botID, tgID)

		menu := app.BuildSubscriptionMenu(ctx)
		text := "⏳ Срок подписки истёк. Продли, чтобы снова пользоваться ботом!"

		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return

	case "active":
		log.Printf("[active] botID=%s tgID=%d → show main menu", botID, tgID)

		// Клавиатура "Остаток минут"
		mainKB := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🕒 Остаток минут"),
			),
		)
		mainKB.ResizeKeyboard = true

		// Отправляем клаву без текста
		msgOut := tgbotapi.NewMessage(chatID, " ") // пробел = телега обновляет клаву
		msgOut.ReplyMarkup = mainKB
		bot.Send(msgOut)

		// Если нажали кнопку
		if msg.Text == "🕒 Остаток минут" {
			log.Printf("[active] botID=%s tgID=%d → ShowVoiceMinutes", botID, tgID)
			app.ShowVoiceMinutesScreen(ctx, botID, bot, tgID, chatID)
			return
		}

		// Обработка типов сообщений
		switch {
		case msg.Voice != nil:
			log.Printf("[active] botID=%s tgID=%d → voice", botID, tgID)
			app.handleVoice(ctx, botID, bot, msg, tgID)

		case len(msg.Photo) > 0:
			log.Printf("[active] botID=%s tgID=%d → photo", botID, tgID)
			app.handlePhoto(ctx, botID, bot, msg, tgID)

		case msg.Text != "":
			log.Printf("[active] botID=%s tgID=%d → text(%q)", botID, tgID, msg.Text)
			app.handleText(ctx, botID, bot, msg, tgID)

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "📎 Отправь текст, голос или фото."))
		}
		return

	default:
		log.Printf("[unknown] botID=%s tgID=%d → status=%s", botID, tgID, status)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Неизвестный статус подписки."))
		return
	}
}

// extractTelegramID — выбирает ID пользователя из Update
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
	if err != nil {
		log.Printf("[voice_check] ❗ Get failed bot=%s tg=%d err=%v", botID, tgID, err)
		return false
	}
	if sub == nil {
		log.Printf("[voice_check] ❗ No subscription bot=%s tg=%d", botID, tgID)
		return false
	}

	log.Printf("[voice_check] status=%s voice_minutes=%d expires=%v",
		sub.Status, sub.VoiceMinutes, sub.ExpiresAt)

	if sub.Status != "active" {
		log.Printf("[voice_check] ❌ Not active")
		return false
	}
	if sub.VoiceMinutes <= 0 {
		log.Printf("[voice_check] ❌ No voice minutes left")
		return false
	}

	log.Printf("[voice_check] ✔ Allowed")
	return true
}

func (app *BotApp) checkImageAllowed(ctx context.Context, botID string, tgID int64) bool {
	return true
}
