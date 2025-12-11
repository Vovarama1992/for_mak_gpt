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

	mainKB := app.BuildMainKeyboard(status)

	switch status {

	// ======================================================
	// NONE → нет подписки, ждём нажатия “Начать урок”
	// ======================================================
	case "none":
		if msg.Text == "🟢 Начать урок" {

			// 1. создаём демо-подписку
			if err := app.SubscriptionService.StartDemo(ctx, botID, tgID); err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании демо-подписки. Попробуй ещё раз."))
				return
			}

			// 2. грузим конфиг бота
			cfg, _ := app.BotsService.Get(ctx, botID)

			// 3. приветственный текст
			welcomeText := strings.TrimSpace(cfg.WelcomeText)
			if welcomeText == "" {
				welcomeText = "Привет! Я — твой AI-репетитор 🤖📚\nВыбери класс, чтобы начать."
			}
			bot.Send(tgbotapi.NewMessage(chatID, welcomeText))

			// 4. приветственное видео (если задано)
			if cfg.WelcomeVideo != "" {
				video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(cfg.WelcomeVideo))
				bot.Send(video)
			}

			// 5. меню выбора класса
			app.ShowClassPicker(ctx, botID, bot, tgID, chatID)
			return
		}

		// любое другое сообщение — мягкое приглашение начать
		welcome := tgbotapi.NewMessage(chatID,
			"Добро пожаловать! Нажми «🟢 Начать урок», чтобы начать обучение.")
		welcome.ReplyMarkup = mainKB
		bot.Send(welcome)
		return

	// ======================================================
	// PENDING → ждём оплаты
	// ======================================================
	case "pending":
		m := tgbotapi.NewMessage(chatID, MsgPending)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return

	// ======================================================
	// EXPIRED → срок вышел
	// ======================================================
	case "expired":
		menu := app.BuildSubscriptionMenu(ctx)
		text := "⏳ Срок подписки истёк. Продли, чтобы снова пользоваться ботом!"
		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return

	// ======================================================
	// ACTIVE → основная логика
	// ======================================================
	case "active":

		// обновляем клавиатуру
		msgOut := tgbotapi.NewMessage(chatID, " ")
		msgOut.ReplyMarkup = mainKB
		bot.Send(msgOut)

		// обработка кнопок
		switch msg.Text {

		case "🟢 Продолжить урок":
			bot.Send(tgbotapi.NewMessage(chatID, "Отправь текст, голос, фото или документ для урока."))
			return

		case "💳 Тарифы":
			menu := app.BuildSubscriptionMenu(ctx)
			t := app.BuildSubscriptionText()
			out := tgbotapi.NewMessage(chatID, t)
			out.ReplyMarkup = menu
			bot.Send(out)
			return

		case "❓ Помощь":
			m := tgbotapi.NewMessage(chatID, "Это репетитор по математике. Отправь задание текстом, голосом, фото или файлом.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
			return

		case "🗑 Очистить историю":
			err := app.RecordService.DeleteUserHistory(ctx, botID, tgID)
			if err != nil {
				m := tgbotapi.NewMessage(chatID, "Не удалось очистить историю.")
				m.ReplyMarkup = mainKB
				bot.Send(m)
				return
			}
			m := tgbotapi.NewMessage(chatID, "История очищена.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
			return
		}

		// обработка типов
		switch {
		case msg.Voice != nil:
			app.handleVoice(ctx, botID, bot, msg, tgID, mainKB)
			return

		case msg.Document != nil:
			if isPDF(msg.Document) {
				app.handlePDF(ctx, botID, bot, msg, tgID, mainKB)
			} else if isWord(msg.Document) {
				app.handleDoc(ctx, botID, bot, msg, tgID, mainKB)
			} else {
				app.handlePhoto(ctx, botID, bot, msg, tgID, mainKB)
			}
			return

		case len(msg.Photo) > 0:
			app.handlePhoto(ctx, botID, bot, msg, tgID, mainKB)
			return

		case msg.Text != "":
			app.handleText(ctx, botID, bot, msg, tgID, mainKB)
			return

		default:
			m := tgbotapi.NewMessage(chatID, "📎 Отправь текст, голос, фото или документ.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
			return
		}

	// ======================================================
	// UNKNOWN
	// ======================================================
	default:
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

func isPDF(doc *tgbotapi.Document) bool {
	name := strings.ToLower(doc.FileName)
	return strings.HasSuffix(name, ".pdf")
}

func isWord(doc *tgbotapi.Document) bool {
	name := strings.ToLower(doc.FileName)
	return strings.HasSuffix(name, ".doc") || strings.HasSuffix(name, ".docx")
}
