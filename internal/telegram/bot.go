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

	tokens := strings.Split(tokensEnv, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		bot, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Printf("failed to init bot: %v", err)
			continue
		}

		app.mu.Lock()
		app.bots[bot.Self.UserName] = bot
		app.mu.Unlock()
		log.Printf("bot ready for send: %s", bot.Self.UserName)
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
	menu.ShowTariffs(ctx, bot, &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: telegramID}}, app.TariffService)
}
