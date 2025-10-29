package telegram

import (
	"context"
	"log"
	"time"

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

		log.Printf("[bot_loop] update received botID=%s tgID=%d", botID, tgID)

		status, err := app.SubscriptionService.GetStatus(ctx, botID, tgID)
		if err != nil {
			log.Printf("[bot_loop] getStatus fail botID=%s tgID=%d err=%v", botID, tgID, err)
			continue
		}

		switch {
		case update.Message != nil:
			app.handleMessage(ctx, botID, bot, update.Message.Chat.ID, tgID, status)

		case update.CallbackQuery != nil:
			app.handleCallback(ctx, botID, bot, update.CallbackQuery, status)
		}
	}
}

// extractTelegramID — достаёт userID из апдейта
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

// handleMessage — обрабатывает любые входящие тексты
func (app *BotApp) handleMessage(ctx context.Context, botID string, bot *tgbotapi.BotAPI, chatID, tgID int64, status string) {
	switch status {
	case "none":
		log.Printf("[bot_loop] no subscription botID=%s tgID=%d → show menu", botID, tgID)
		menu := app.BuildSubscriptionMenu(ctx)
		msg := tgbotapi.NewMessage(chatID, MsgNoSubscription)
		msg.ReplyMarkup = menu
		bot.Send(msg)

	case "pending":
		log.Printf("[bot_loop] pending botID=%s tgID=%d", botID, tgID)
		msg := tgbotapi.NewMessage(chatID, MsgPending)
		bot.Send(msg)

	case "active":
		log.Printf("[bot_loop] active botID=%s tgID=%d", botID, tgID)
		msg := tgbotapi.NewMessage(chatID, MsgActive)
		bot.Send(msg)

	default:
		log.Printf("[bot_loop] unknown status=%s botID=%s tgID=%d", status, botID, tgID)
	}
}

// handleCallback — выбор тарифа из меню
func (app *BotApp) handleCallback(ctx context.Context, botID string, bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, status string) {
	tgID := cb.From.ID
	chatID := cb.Message.Chat.ID
	log.Printf("[bot_loop] callback botID=%s tgID=%d data=%s", botID, tgID, cb.Data)

	switch status {
	case "none":
		err := app.SubscriptionService.Create(ctx, botID, tgID, cb.Data)
		if err != nil {
			log.Printf("[bot_loop] create pending fail botID=%s tgID=%d: %v", botID, tgID, err)
			return
		}
		bot.Request(tgbotapi.NewCallback(cb.ID, "Заявка принята"))
		msg := tgbotapi.NewMessage(chatID, MsgAccepted)
		bot.Send(msg)

	case "pending", "active":
		bot.Request(tgbotapi.NewCallback(cb.ID, "Подписка уже оформлена"))
		msg := tgbotapi.NewMessage(chatID, MsgAlreadySubscribed)
		bot.Send(msg)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("[bot_loop] init %s", time.Now().Format(time.RFC3339))
}
