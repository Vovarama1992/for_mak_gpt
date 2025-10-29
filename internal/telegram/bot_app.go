package telegram

import (
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

// InitBots — инициализирует всех ботов из BOT_TOKENS
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

		log.Printf("[InitBots] bot ready: @%s (bot_id=%s)", bot.Self.UserName, botID)
	}

	return nil
}
