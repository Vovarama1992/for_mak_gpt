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

		// =========================================
		// ФАКТ ВЗАИМОДЕЙСТВИЯ С БОТОМ (САМЫЙ ВАЖНЫЙ ЛОГ)
		// =========================================
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

		// =========================================
		// ДАЛЬШЕ — ТВОЯ СУЩЕСТВУЮЩАЯ ЛОГИКА
		// =========================================

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

	log.Printf("[sub-check] botID=%s tgID=%d → status=%s", botID, tgID, status)

	mainKB := app.BuildMainKeyboard(status)

	// ======================================================
	// ГЛОБАЛЬНЫЕ КОМАНДЫ (НЕ ЗАВИСЯТ ОТ СТАТУСА)
	// ======================================================

	switch msg.Text {

	case "❓ Помощь":
		if app.adminBotUsername == "" {
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"Поддержка временно недоступна.",
			))
			return
		}

		url := "https://t.me/" + app.adminBotUsername + "?start=support"

		m := tgbotapi.NewMessage(
			chatID,
			"🆘 Чтобы написать в поддержку, нажми кнопку ниже:",
		)
		m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"✉️ Написать в поддержку",
					url,
				),
			),
		)

		bot.Send(m)
		return

	case "💳 Тарифы":
		menu := app.BuildSubscriptionMenu(ctx, botID)
		text := app.BuildSubscriptionText(ctx, botID)

		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	// ======================================================
	// КЛЮЧЕВАЯ ЧАСТЬ — ПРОВЕРКА КЛАССА
	// ======================================================

	userClass, _ := app.ClassService.GetUserClass(ctx, botID, tgID)

	// ======================================================
	// STATUS FLOW
	// ======================================================

	switch status {

	case "none":

		log.Printf("[flow:none] enter bot=%s tg=%d userClass=%v",
			botID, tgID, userClass != nil,
		)

		if userClass != nil {

			trialTariff, err := app.TariffService.GetTrial(ctx, botID)
			if err != nil {
				log.Printf("[flow:none] GetTrial error: %v", err)
			} else if trialTariff == nil {
				log.Printf("[flow:none] trialTariff = nil")
			} else {
				log.Printf("[flow:none] trialTariff found code=%s", trialTariff.Code)

				err := app.SubscriptionService.ActivateTrial(
					ctx, botID, tgID, trialTariff.Code,
				)
				if err != nil {
					log.Printf("[flow:none] ActivateTrial error: %v", err)
				}
			}

			newStatus, err := app.SubscriptionService.GetStatus(ctx, botID, tgID)
			if err == nil {
				status = newStatus
			}

			if status != "active" {
				menu := app.BuildSubscriptionMenu(ctx, botID)
				out := tgbotapi.NewMessage(
					chatID,
					"⛔ Подписка не активна. Оформи подписку, чтобы продолжить обучение.",
				)
				out.ReplyMarkup = menu
				bot.Send(out)
				return
			}

			msgOut := tgbotapi.NewMessage(chatID, " ")
			msgOut.ReplyMarkup = app.BuildMainKeyboard("active")
			bot.Send(msgOut)

			switch {
			case msg.Voice != nil:
				app.handleVoice(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
			case msg.Document != nil:
				if isPDF(msg.Document) {
					app.handlePDF(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
				} else if isWord(msg.Document) {
					app.handleDoc(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
				} else {
					app.handlePhoto(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
				}
			case len(msg.Photo) > 0:
				app.handlePhoto(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
			case msg.Text != "":
				app.handleText(ctx, botID, bot, msg, tgID, app.BuildMainKeyboard("active"))
			}
			return
		}

		if msg.Text == "🟢 Начать урок" {

			trialTariff, err := app.TariffService.GetTrial(ctx, botID)
			if err != nil || trialTariff == nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Пробный тариф не настроен. Обратись к администратору.",
				))
				return
			}

			if err := app.SubscriptionService.ActivateTrial(
				ctx, botID, tgID, trialTariff.Code,
			); err != nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Ошибка при активации пробного периода.",
				))
				return
			}

			cfg, _ := app.BotsService.Get(ctx, botID)

			welcomeText := "Привет! Я — твой AI-репетитор 🤖📚\nВыбери класс, чтобы начать."
			if cfg != nil && cfg.WelcomeText != nil {
				welcomeText = strings.TrimSpace(*cfg.WelcomeText)
			}

			bot.Send(tgbotapi.NewMessage(chatID, welcomeText))

			if cfg != nil && cfg.WelcomeVideo != nil && *cfg.WelcomeVideo != "" {
				video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(*cfg.WelcomeVideo))
				bot.Send(video)
			}

			app.ShowClassPicker(ctx, botID, bot, tgID, chatID)
			return
		}

		welcome := tgbotapi.NewMessage(
			chatID,
			"Добро пожаловать! Нажми «🟢 Начать урок», чтобы начать обучение.",
		)
		welcome.ReplyMarkup = mainKB
		bot.Send(welcome)
		return

	case "pending":
		m := tgbotapi.NewMessage(chatID, MsgPending)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return

	case "expired":
		menu := app.BuildSubscriptionMenu(ctx, botID)
		out := tgbotapi.NewMessage(
			chatID,
			"⛔ Подписка истекла.\nОформи подписку, чтобы продолжить обучение.",
		)
		out.ReplyMarkup = menu
		bot.Send(out)
		return

	case "active":

		msgOut := tgbotapi.NewMessage(chatID, " ")
		msgOut.ReplyMarkup = mainKB
		bot.Send(msgOut)

		switch msg.Text {

		case "🟢 Продолжить урок":
			cfg, _ := app.BotsService.Get(ctx, botID)

			text := "Отправь текст, голос, фото или документ для урока."
			if cfg != nil && cfg.AfterContinueText != nil {
				if t := strings.TrimSpace(*cfg.AfterContinueText); t != "" {
					text = t
				}
			}

			bot.Send(tgbotapi.NewMessage(chatID, text))
			return

		case "🗑 Очистить историю":
			if err := app.RecordService.DeleteUserHistory(ctx, botID, tgID); err != nil {
				m := tgbotapi.NewMessage(chatID, "Не удалось очистить историю.")
				m.ReplyMarkup = mainKB
				bot.Send(m)
				return
			}
			m := tgbotapi.NewMessage(chatID, "История очищена.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
			return

		case "🧹 Сбросить настройки":
			if err := app.UserService.ResetUserSettings(ctx, botID, tgID); err != nil {
				m := tgbotapi.NewMessage(chatID, "Не удалось сбросить настройки.")
				m.ReplyMarkup = mainKB
				bot.Send(m)
				return
			}

			m := tgbotapi.NewMessage(chatID, "Настройки сброшены. Можешь начать заново.")
			m.ReplyMarkup = app.BuildMainKeyboard("none")
			bot.Send(m)
			return

		case "📦 Пакеты минут":
			menu := app.BuildMinutePackagesMenu(ctx, botID)
			out := tgbotapi.NewMessage(chatID, "Выбери пакет минут:")
			out.ReplyMarkup = menu
			bot.Send(out)
			return
		}

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
		case msg.Text != "":
			app.handleText(ctx, botID, bot, msg, tgID, mainKB)
		default:
			m := tgbotapi.NewMessage(chatID, "📎 Отправь текст, голос, фото или документ.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
		}
		return
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
