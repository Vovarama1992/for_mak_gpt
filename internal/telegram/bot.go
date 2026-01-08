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

	log.Printf("[sub-check] botID=%s tgID=%d → status=%s", botID, tgID, status)

	mainKB := app.BuildMainKeyboard(status)
	textLower := strings.ToLower(msg.Text)

	// =====================================================
	// ГЛОБАЛЬНЫЕ КОМАНДЫ (НЕ ЗАВИСЯТ ОТ STATUS)
	// =====================================================
	switch {

	case strings.Contains(textLower, "помощ"):
		if app.adminBotUsername == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Поддержка временно недоступна."))
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

	case strings.Contains(textLower, "тариф"):
		menu := app.BuildSubscriptionMenu(ctx, botID)
		text := app.BuildSubscriptionText(ctx, botID)

		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return

	case strings.Contains(textLower, "минут"):
		menu := app.BuildMinutePackagesMenu(ctx, botID)
		out := tgbotapi.NewMessage(chatID, "Выбери пакет минут:")
		out.ReplyMarkup = menu
		bot.Send(out)
		return
	}

	userClass, _ := app.ClassService.GetUserClass(ctx, botID, tgID)

	// =====================================================
	// ОСНОВНОЙ FLOW ПО STATUS
	// =====================================================
	switch status {

	case "none":

		log.Printf("[flow:none] enter bot=%s tg=%d userClass=%v",
			botID, tgID, userClass != nil,
		)

		if botID == "assistant" {

			if strings.Contains(textLower, "начать") {

				trialTariff, err := app.TariffService.GetTrial(ctx, botID)
				if err != nil || trialTariff == nil {
					bot.Send(tgbotapi.NewMessage(
						chatID,
						"Пробный тариф не настроен. Обратись к администратору.",
					))
					return
				}

				_ = app.SubscriptionService.ActivateTrial(
					ctx, botID, tgID, trialTariff.Code,
				)

				cfg, _ := app.BotsService.Get(ctx, botID)

				welcomeText := "Привет! Я — твой AI-ассистент 🤖"
				if cfg != nil && cfg.WelcomeText != nil {
					welcomeText = strings.TrimSpace(*cfg.WelcomeText)
				}

				bot.Send(tgbotapi.NewMessage(chatID, welcomeText))

				msgOut := tgbotapi.NewMessage(chatID, " ")
				msgOut.ReplyMarkup = app.BuildMainKeyboard("active")
				bot.Send(msgOut)
				return
			}

			welcome := tgbotapi.NewMessage(
				chatID,
				"Добро пожаловать! Нажми «🟢 Начать урок», чтобы начать.",
			)
			welcome.ReplyMarkup = mainKB
			bot.Send(welcome)
			return
		}

		if userClass != nil {

			trialTariff, _ := app.TariffService.GetTrial(ctx, botID)
			if trialTariff != nil {
				_ = app.SubscriptionService.ActivateTrial(
					ctx, botID, tgID, trialTariff.Code,
				)
			}

			newStatus, _ := app.SubscriptionService.GetStatus(ctx, botID, tgID)
			if newStatus != "" {
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
			return
		}

		if strings.Contains(textLower, "начать") {

			trialTariff, err := app.TariffService.GetTrial(ctx, botID)
			if err != nil || trialTariff == nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Пробный тариф не настроен. Обратись к администратору.",
				))
				return
			}

			_ = app.SubscriptionService.ActivateTrial(
				ctx, botID, tgID, trialTariff.Code,
			)

			cfg, _ := app.BotsService.Get(ctx, botID)

			welcomeText := "Привет! Я — твой AI-репетитор 🤖📚\nВыбери класс, чтобы начать."
			if cfg != nil && cfg.WelcomeText != nil {
				welcomeText = strings.TrimSpace(*cfg.WelcomeText)
			}

			bot.Send(tgbotapi.NewMessage(chatID, welcomeText))
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

		switch {

		case strings.Contains(textLower, "продолж"):
			cfg, _ := app.BotsService.Get(ctx, botID)

			text := "Отправь текст, голос, фото или документ для урока."
			if cfg != nil && cfg.AfterContinueText != nil {
				if t := strings.TrimSpace(*cfg.AfterContinueText); t != "" {
					text = t
				}
			}

			bot.Send(tgbotapi.NewMessage(chatID, text))
			return

		case strings.Contains(textLower, "очист"):
			_ = app.RecordService.DeleteUserHistory(ctx, botID, tgID)
			m := tgbotapi.NewMessage(chatID, "История очищена.")
			m.ReplyMarkup = mainKB
			bot.Send(m)
			return

		case strings.Contains(textLower, "сброс"):
			_ = app.UserService.ResetUserSettings(ctx, botID, tgID)
			m := tgbotapi.NewMessage(chatID, "Настройки сброшены. Можешь начать заново.")
			m.ReplyMarkup = app.BuildMainKeyboard("none")
			bot.Send(m)
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
