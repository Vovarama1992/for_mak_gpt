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

	// ВСЕГДА сразу отвечаем Telegram
	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	log.Printf("[callback] botID=%s tgID=%d data=%s", botID, tgID, data)

	// ---------------------------
	// 1) Покупка минут — показать список
	// ---------------------------
	if data == "buy_voice" {
		menu := app.BuildMinutePackagesMenu(ctx)

		edit := tgbotapi.NewEditMessageText(
			chatID,
			cb.Message.MessageID,
			"Выбери пакет минут:",
		)
		edit.ReplyMarkup = &menu
		bot.Request(edit)
		return
	}

	// ---------------------------
	// 2) Выбор класса
	// ---------------------------
	if strings.HasPrefix(data, "set_class_") {
		idStr := strings.TrimPrefix(data, "set_class_")
		classID, _ := strconv.Atoi(idStr)

		if err := app.ClassService.SetUserClass(ctx, botID, tgID, classID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось установить класс"))
			return
		}

		// УБИРАЕМ inline-меню корректно
		edit := tgbotapi.NewEditMessageReplyMarkup(
			chatID,
			cb.Message.MessageID,
			tgbotapi.InlineKeyboardMarkup{},
		)
		bot.Request(edit)

		bot.Send(tgbotapi.NewMessage(chatID, "Класс обновлён"))
		return
	}

	// ---------------------------
	// 3) Пакеты минут
	// ---------------------------
	if strings.HasPrefix(data, "pkg_") {
		idStr := strings.TrimPrefix(data, "pkg_")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		pkg, err := app.MinutePackageService.GetByID(ctx, id)
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

		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("🔄 Для оплаты перейди по ссылке:\n%s", payURL)))
		return
	}

	// ---------------------------
	// 4) Подписки
	// ---------------------------
	switch status {

	case "none":
		paymentURL, err := app.SubscriptionService.Create(ctx, botID, tgID, data)
		if err != nil {
			app.ErrorNotify.Notify(ctx, botID, err, "Ошибка создания подписки")
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось создать оплату."))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Заявка принята!\n%s", paymentURL)))
		return

	case "pending":
		bot.Send(tgbotapi.NewMessage(chatID, "⏳ Ожидается подтверждение оплаты."))
		return

	case "active":
		bot.Send(tgbotapi.NewMessage(chatID, MsgAlreadySubscribed))
		return

	default:
		err := fmt.Errorf("unexpected status '%s' for callback '%s'", status, data)
		app.ErrorNotify.Notify(ctx, botID, err, "Неожиданный callback")
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Произошла ошибка."))
		return
	}
}
