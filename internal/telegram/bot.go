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
		log.Println("no BOT_TOKENS provided")
		return nil
	}

	// порядок: ai_assistant=token1,copy_assistant=token2
	pairs := strings.Split(tokensEnv, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			log.Printf("invalid BOT_TOKENS pair: %s", pair)
			continue
		}

		botID := parts[0]
		token := parts[1]

		bot, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Printf("failed to init bot %s: %v", botID, err)
			continue
		}

		app.mu.Lock()
		app.bots[botID] = bot
		app.mu.Unlock()

		log.Printf("bot ready for send: %s (bot_id=%s)", bot.Self.UserName, botID)
	}
	return nil
}

// CheckSubscriptionAndShowMenu проверяет подписку и показывает меню, если её нет
func (app *BotApp) CheckSubscriptionAndShowMenu(ctx context.Context, botID string, telegramID int64) {
	app.mu.RLock()
	bot := app.bots[botID]
	app.mu.RUnlock()
	if bot == nil {
		log.Printf("bot not found for id: %s", botID)
		return
	}

	status, err := app.SubscriptionService.GetStatus(ctx, botID, telegramID)
	if err != nil {
		log.Printf("failed to get subscription status: %v", err)
		return
	}

	if status == "active" || status == "pending" {
		return // подписка есть — ничего не показываем
	}

	menu := NewMenu()
	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: telegramID}}
	menu.ShowTariffs(ctx, botID, bot, msg, app.TariffService)
}
