package telegram

import (
	"context"
	"fmt"
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

	// ======================================================
	// ADMIN HELP MODE → ONE MESSAGE REPLY
	// ======================================================
	if isAdmin(tgID) {
		if ctxHelp, ok := app.adminHelpMode[tgID]; ok {
			reply := "💬 Ответ от поддержки:\n\n" + msg.Text

			bot.Send(tgbotapi.NewMessage(
				ctxHelp.UserID,
				reply,
			))

			delete(app.adminHelpMode, tgID)

			bot.Send(tgbotapi.NewMessage(
				chatID,
				"✅ Ответ отправлен пользователю.",
			))
			return
		}
	}

	log.Printf("[sub-check] botID=%s tgID=%d → status=%s", botID, tgID, status)

	mainKB := app.BuildMainKeyboard(status)

	// ======================================================
	// USER HELP MODE
	// ======================================================
	if app.helpMode[botID] != nil && app.helpMode[botID][tgID] {

		if msg.Text == "⬅️ Назад" {
			delete(app.helpMode[botID], tgID)

			m := tgbotapi.NewMessage(
				chatID,
				"Ты вышел из режима помощи.",
			)
			m.ReplyMarkup = app.BuildMainKeyboard(status)
			bot.Send(m)
			return
		}

		text := "🆘 Помощь\n" +
			"Bot: " + botID + "\n" +
			"UserID: " + fmt.Sprintf("%d", tgID) + "\n\n" +
			msg.Text

		admins := []int64{
			1139929360,
			6789440333,
		}

		for _, adminID := range admins {
			bot.Send(tgbotapi.NewMessage(adminID, text))

			// ВКЛЮЧАЕМ ADMIN HELP MODE
			app.adminHelpMode[adminID] = &AdminHelpContext{
				BotID:  botID,
				UserID: tgID,
			}
		}

		bot.Send(tgbotapi.NewMessage(
			chatID,
			"Сообщение отправлено администратору. Ожидай ответа.",
		))
		return
	}

	// ======================================================
	// STATUS FLOW
	// ======================================================
	switch status {

	case "none":
		if msg.Text == "🟢 Начать урок" {

			trialTariff, err := app.TariffService.GetTrial(ctx)
			if err != nil || trialTariff == nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Пробный тариф не настроен. Обратись к администратору.",
				))
				return
			}

			if err := app.SubscriptionService.ActivateTrial(
				ctx,
				botID,
				tgID,
				trialTariff.Code,
			); err != nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Ошибка при активации пробного периода.",
				))
				return
			}

			cfg, err := app.BotsService.Get(ctx, botID)
			if err != nil {
				log.Printf("[welcome] failed to load bot config: %v", err)
			}

			var welcomeText string
			if cfg != nil && cfg.WelcomeText != nil {
				welcomeText = strings.TrimSpace(*cfg.WelcomeText)
			}
			if welcomeText == "" {
				welcomeText = "Привет! Я — твой AI-репетитор 🤖📚\nВыбери класс, чтобы начать."
			}
			bot.Send(tgbotapi.NewMessage(chatID, welcomeText))

			if cfg != nil && cfg.WelcomeVideo != nil && *cfg.WelcomeVideo != "" {
				video := tgbotapi.NewVideo(
					chatID,
					tgbotapi.FileURL(*cfg.WelcomeVideo),
				)
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
		menu := app.BuildSubscriptionMenu(ctx)
		text := "⏳ Срок подписки истёк. Продли, чтобы снова пользоваться ботом!"
		out := tgbotapi.NewMessage(chatID, text)
		out.ReplyMarkup = menu
		bot.Send(out)
		return

	case "active":

		msgOut := tgbotapi.NewMessage(chatID, " ")
		msgOut.ReplyMarkup = mainKB
		bot.Send(msgOut)

		switch msg.Text {

		case "🟢 Продолжить урок":
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"Отправь текст, голос, фото или документ для урока.",
			))
			return

		case "💳 Тарифы":
			menu := app.BuildSubscriptionMenu(ctx)
			text := app.BuildSubscriptionText()
			out := tgbotapi.NewMessage(chatID, text)
			out.ReplyMarkup = menu
			bot.Send(out)
			return

		case "❓ Помощь":
			if app.helpMode[botID] == nil {
				app.helpMode[botID] = make(map[int64]bool)
			}
			app.helpMode[botID][tgID] = true

			m := tgbotapi.NewMessage(
				chatID,
				"🆘 Напиши сообщение — его получит администратор.\nЧтобы выйти, нажми «Назад».",
			)
			m.ReplyMarkup = helpKeyboard()
			bot.Send(m)
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

			m := tgbotapi.NewMessage(
				chatID,
				"Настройки сброшены. Можешь начать заново.",
			)
			m.ReplyMarkup = app.BuildMainKeyboard("none")
			bot.Send(m)
			return
		}

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

	default:
		bot.Send(tgbotapi.NewMessage(
			chatID,
			"⚠️ Неизвестный статус подписки.",
		))
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

func isAdmin(id int64) bool {
	return id == 1139929360 || id == 6789440333
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
