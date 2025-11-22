package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleCallback(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	cb *tgbotapi.CallbackQuery, status string) {

	tgID := cb.From.ID
	chatID := cb.Message.Chat.ID
	data := cb.Data

	log.Printf("[callback] botID=%s tgID=%d data=%s", botID, tgID, data)

	switch status {

	case "none":
		// пользователь выбирает тариф
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
		// платеж уже создан, но не оплачен
		bot.Request(tgbotapi.NewCallback(cb.ID, "Платёж уже создан"))
		bot.Send(tgbotapi.NewMessage(chatID, "⏳ Ожидается подтверждение оплаты."))
		return

	case "active":
		// активный подписчик кликает на кнопки тарифов
		bot.Request(tgbotapi.NewCallback(cb.ID, "Уже подписан"))
		bot.Send(tgbotapi.NewMessage(chatID, MsgAlreadySubscribed))
		return

	default:
		// неизвестный статус (вообще не должен случаться)
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
