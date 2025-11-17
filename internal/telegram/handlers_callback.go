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

	log.Printf("[callback] botID=%s tgID=%d data=%s", botID, tgID, cb.Data)

	switch status {
	case "none":
		paymentURL, err := app.SubscriptionService.Create(ctx, botID, tgID, cb.Data)
		if err != nil {
			log.Printf("[callback] create payment fail: %v", err)
			bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка оформления"))
			bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка при создании оплаты, попробуй позже."))
			return
		}

		bot.Request(tgbotapi.NewCallback(cb.ID, "Заявка принята"))
		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Заявка принята!\nДля оплаты перейди по ссылке:\n%s", paymentURL)))

	case "pending", "active":
		bot.Request(tgbotapi.NewCallback(cb.ID, "Уже подписан"))
		bot.Send(tgbotapi.NewMessage(chatID, MsgAlreadySubscribed))
	}
}
