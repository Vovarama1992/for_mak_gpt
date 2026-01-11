package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleCallback(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	cb *tgbotapi.CallbackQuery,
	status string,
) {
	tgID := cb.From.ID
	chatID := cb.Message.Chat.ID
	data := cb.Data

	// всегда отвечаем Telegram
	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	log.Printf("[callback] botID=%s tgID=%d data=%s", botID, tgID, data)

	// ---------------------------
	// 2) Выбор класса
	// ---------------------------
	if strings.HasPrefix(data, "set_class_") {
		idStr := strings.TrimPrefix(data, "set_class_")
		classID, err := strconv.Atoi(idStr)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Некорректный класс"))
			return
		}

		// сохранить класс
		if err := app.ClassService.SetUserClass(ctx, botID, tgID, classID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось установить класс"))
			return
		}

		// ДОСТАЁМ класс
		class, err := app.ClassService.GetClassByID(ctx, botID, classID)
		if err != nil || class == nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Класс выбран"))
			return
		}

		// убираем inline
		bot.Request(tgbotapi.NewEditMessageReplyMarkup(
			chatID,
			cb.Message.MessageID,
			tgbotapi.InlineKeyboardMarkup{},
		))

		// точная отбивка
		m := tgbotapi.NewMessage(
			chatID,
			fmt.Sprintf("Выбран %s. Можем начинать 👍", class.Grade),
		)
		m.ReplyMarkup = app.BuildMainKeyboard(botID, "active")
		bot.Send(m)
		return
	}
	// ---------------------------
	// 3) Пакеты минут
	// ---------------------------
	if strings.HasPrefix(data, "pkg_") {
		idStr := strings.TrimPrefix(data, "pkg_")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		pkg, err := app.MinutePackageService.GetByID(ctx, botID, id)
		if err != nil || pkg == nil || !pkg.Active {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Пакет недоступен."))
			return
		}

		payURL, err := app.MinutePackageService.CreatePayment(ctx, botID, tgID, pkg.ID)
		if err != nil {
			app.ErrorNotify.Notify(ctx, botID, err, "Ошибка создания платежа")
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось создать оплату."))
			return
		}

		msg := tgbotapi.NewMessage(
			chatID,
			fmt.Sprintf("🔄 Для оплаты перейди по ссылке:\n%s", payURL),
		)
		msg.ReplyMarkup = app.BuildMainKeyboard(botID, status)
		bot.Send(msg)
		return
	}

	// ---------------------------
	// 4) Подписка (без предпросмотров)
	// ---------------------------
	if strings.HasPrefix(data, "sub:") {
		planCode := strings.TrimPrefix(data, "sub:")

		switch status {
		case "active":
			bot.Send(tgbotapi.NewMessage(chatID, MsgAlreadySubscribed))
			return

		case "pending":
			bot.Send(tgbotapi.NewMessage(chatID, "⏳ Ожидается подтверждение оплаты."))
			return

		case "none", "expired":
			paymentURL, err := app.SubscriptionService.Create(ctx, botID, tgID, planCode)
			if err != nil {
				app.ErrorNotify.Notify(ctx, botID, err, "Ошибка создания подписки")
				bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось создать оплату."))
				return
			}

			msg := tgbotapi.NewMessage(
				chatID,
				fmt.Sprintf("✅ Ссылка на оплату\n%s", paymentURL),
			)
			msg.ReplyMarkup = app.BuildMainKeyboard(botID, status)
			bot.Send(msg)
			return
		}
	}

	// ---------------------------
	// 5) Активация TRIAL
	// ---------------------------
	if data == "activate_trial" {

		trial, err := app.TariffService.GetTrial(ctx, botID)
		if err != nil || trial == nil {
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"❗ Пробный тариф недоступен.",
			))
			return
		}

		// просто активируем — UI уже проверил, что trial не был
		if err := app.SubscriptionService.ActivateTrial(
			ctx,
			botID,
			tgID,
			trial.Code,
		); err != nil {
			app.ErrorNotify.Notify(ctx, botID, err, "Ошибка активации trial")
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"⚠️ Не удалось активировать пробный тариф.",
			))
			return
		}

		msg := tgbotapi.NewMessage(
			chatID,
			fmt.Sprintf(
				"✅ Пробный тариф активирован\n⏳ Дней: %d\n🎧 Голосовых минут: %.0f",
				trial.DurationMinutes/(60*24),
				trial.VoiceMinutes,
			),
		)
		msg.ReplyMarkup = app.BuildMainKeyboard(botID, "active")
		bot.Send(msg)
		return
	}

	// ---------------------------
	// неизвестный callback
	// ---------------------------
	err := fmt.Errorf("unknown callback data: %s", data)
	app.ErrorNotify.Notify(ctx, botID, err, "Неизвестный callback")
	bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Произошла ошибка."))
}
