package telegram

import (
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RunListeners — запускает обработку апдейтов у всех ботов
func (app *BotApp) RunListeners(ctx context.Context) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for botID, bot := range app.bots {
		go func(botID string, bot *tgbotapi.BotAPI) {
			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60
			updates := bot.GetUpdatesChan(u)

			log.Printf("[RunListeners] started for bot_id=%s (@%s)", botID, bot.Self.UserName)

			for update := range updates {
				// --- callback-клик по тарифу ---
				if update.CallbackQuery != nil {
					data := update.CallbackQuery.Data
					if strings.HasPrefix(data, "subscribe:") {
						planCode := strings.TrimPrefix(data, "subscribe:")
						tid := update.CallbackQuery.Message.Chat.ID

						err := app.SubscriptionService.Create(ctx, botID, tid, planCode)
						if err != nil {
							log.Printf("[RunListeners] failed to create subscription: %v", err)
							bot.Send(tgbotapi.NewMessage(tid, "⚠️ Ошибка при оформлении подписки. Попробуйте позже."))
							continue
						}

						bot.Send(tgbotapi.NewMessage(tid, "📝 Заявка на подписку создана и обрабатывается."))
					}
					continue
				}

				// --- текстовые сообщения ---
				if update.Message == nil {
					continue
				}
				tid := update.Message.Chat.ID
				text := update.Message.Text

				log.Printf("[RunListeners] message from user=%d bot=%s: %q", tid, botID, text)

				if text == "/start" {
					go app.CheckSubscriptionAndShowMenu(ctx, botID, tid)
					continue
				}

				status, _ := app.SubscriptionService.GetStatus(ctx, botID, tid)
				if status == "pending" {
					bot.Send(tgbotapi.NewMessage(tid, "⏳ Ваша заявка на подписку уже обрабатывается."))
				} else if status == "active" {
					bot.Send(tgbotapi.NewMessage(tid, "✅ Ваша подписка активна. Бот временно на обновлении 🚧"))
				} else {
					go app.CheckSubscriptionAndShowMenu(ctx, botID, tid)
				}
			}
		}(botID, bot)
	}
}
