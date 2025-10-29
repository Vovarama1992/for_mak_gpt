package telegram

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotApp struct {
	SubscriptionService ports.SubscriptionService
	TariffService       ports.TariffService
	bots                map[string]*tgbotapi.BotAPI
	mu                  sync.RWMutex
}

func (app *BotApp) InitBots() error {
	app.bots = make(map[string]*tgbotapi.BotAPI)
	tokensEnv := os.Getenv("BOT_TOKENS")
	if tokensEnv == "" {
		log.Println("[InitBots] no BOT_TOKENS provided")
		return nil
	}

	pairs := strings.Split(tokensEnv, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			log.Printf("[InitBots] invalid BOT_TOKENS pair: %s", pair)
			continue
		}

		botID := parts[0]
		token := parts[1]

		bot, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Printf("[InitBots] failed to init bot %s: %v", botID, err)
			continue
		}

		app.mu.Lock()
		app.bots[botID] = bot
		app.mu.Unlock()

		log.Printf("[InitBots] bot ready for send: @%s (bot_id=%s)", bot.Self.UserName, botID)
	}
	return nil
}

// CheckSubscriptionAndShowMenu проверяет подписку и показывает меню, если её нет
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
		log.Printf("[CheckSubscriptionAndShowMenu] failed to get subscription status: %v", err)
		return
	}

	log.Printf("[CheckSubscriptionAndShowMenu] subscription status=%s", status)

	if status == "active" || status == "pending" {
		log.Printf("[CheckSubscriptionAndShowMenu] subscription already active or pending, skip showing menu")
		return
	}

	menu := NewMenu()
	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: telegramID}}

	log.Printf("[CheckSubscriptionAndShowMenu] showing tariffs menu for user=%d via bot=%s", telegramID, botID)
	menu.ShowTariffs(ctx, botID, bot, msg, app.TariffService)

	log.Printf("[CheckSubscriptionAndShowMenu] menu.ShowTariffs finished for user=%d", telegramID)
}
