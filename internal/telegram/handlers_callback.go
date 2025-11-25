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

	log.Printf("[callback] botID=%s tgID=%d data=%s", botID, tgID, data)

	// ---------------------------
	// 1) Покупка минут — показать список
	// ---------------------------
	if data == "buy_voice" {
		menu := app.BuildMinutePackagesMenu(ctx)
		msg := tgbotapi.NewMessage(chatID, "Выбери пакет минут:")
		msg.ReplyMarkup = menu
		bot.Send(msg)
		return
	}

	// ---------------------------
	// 2) Пользователь выбрал конкретный пакет минут: pkg_{id}
	// ---------------------------
	if strings.HasPrefix(data, "pkg_") {
		idStr := strings.TrimPrefix(data, "pkg_")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		pkg, err := app.MinutePackageService.GetByID(ctx, id)
		if err != nil || pkg == nil || !pkg.Active {
			bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка"))
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Пакет недоступен."))
			return
		}

		// создаём платёж (метод появится позже)
		payURL, err := app.MinutePackageService.CreatePayment(ctx, botID, tgID, pkg.ID)
		if err != nil {
			app.ErrorNotify.Notify(ctx, botID, err,
				fmt.Sprintf("Ошибка создания платежа за пакет минут (%d)", id))

			bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка"))
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось создать оплату. Попробуй позже."))
			return
		}

		bot.Request(tgbotapi.NewCallback(cb.ID, "Открываю оплату"))
		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("🔄 Для оплаты перейди по ссылке:\n%s", payURL)))
		return
	}

	// ---------------------------
	// 3) Подписки (старые тарифы)
	// ---------------------------
	switch status {

	case "none":
		paymentURL, err := app.SubscriptionService.Create(ctx, botID, tgID, data)
		if err != nil {
			log.Printf("[callback] create payment fail: %v", err)

			app.ErrorNotify.Notify(
				ctx,
				botID,
				err,
				fmt.Sprintf("❗ Ошибка создания платежа\n\nБот: %s\nПользователь: %d\nТариф: %s",
					botID, tgID, data),
			)

			bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка оформления"))
			bot.Send(tgbotapi.NewMessage(chatID,
				"⚠️ Не удалось создать оплату. Попробуй позже."))
			return
		}

		bot.Request(tgbotapi.NewCallback(cb.ID, "Заявка принята"))
		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Заявка принята!\nДля оплаты перейди по ссылке:\n%s", paymentURL)))
		return

	case "pending":
		bot.Request(tgbotapi.NewCallback(cb.ID, "Платёж уже создан"))
		bot.Send(tgbotapi.NewMessage(chatID, "⏳ Ожидается подтверждение оплаты."))
		return

	case "active":
		bot.Request(tgbotapi.NewCallback(cb.ID, "Уже подписан"))
		bot.Send(tgbotapi.NewMessage(chatID, MsgAlreadySubscribed))
		return

	default:
		err := fmt.Errorf("unexpected status '%s' for callback '%s'", status, data)
		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf("❗ Неожиданный callback\n\nБот: %s\nПользователь: %d\nStatus: %s\nData: %s",
				botID, tgID, status, data),
		)

		bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка"))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Произошла ошибка."))
		return
	}
}
