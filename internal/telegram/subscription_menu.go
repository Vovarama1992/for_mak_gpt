package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckSubscriptionAndShowMenu — проверяет подписку и показывает меню, если её нет
func (app *BotApp) CheckSubscriptionAndShowMenu(ctx context.Context, botID string, telegramID int64) {
	log.Printf("[CheckSubscriptionAndShowMenu] start: bot_id=%s, telegram_id=%d", botID, telegramID)

	app.mu.RLock()
	bot := app.bots[botID]
	app.mu.RUnlock()

	if bot == nil {
		log.Printf("[CheckSubscriptionAndShowMenu] bot not found for id: %s", botID)
		return
	}

	status, err := app.SubscriptionService.GetStatus(ctx, botID, telegramID)
	if err != nil {
		log.Printf("[CheckSubscriptionAndShowMenu] get status error: %v", err)
		return
	}

	switch status {
	case "active":
		bot.Send(tgbotapi.NewMessage(telegramID, "✅ Ваша подписка активна. Бот временно на обновлении 🚧"))
		return
	case "pending":
		bot.Send(tgbotapi.NewMessage(telegramID, "⏳ Ваша заявка на подписку уже обрабатывается."))
		return
	default:
		menu := NewMenu()
		msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: telegramID}}
		log.Printf("[CheckSubscriptionAndShowMenu] showing tariff menu for user %d via %s", telegramID, botID)
		menu.ShowTariffs(ctx, botID, bot, msg, app.TariffService)
	}
}
