package telegram

import (
	"log"
	"os"
	"strings"

	"github.com/Vovarama1992/make_ziper/internal/ai"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotApp struct {
	SubscriptionService ports.SubscriptionService
	TariffService       ports.TariffService
	AiService           ai.Service
	bots                map[string]*tgbotapi.BotAPI
}

func (app *BotApp) InitBots() error {
	app.bots = make(map[string]*tgbotapi.BotAPI)

	env := os.Getenv("BOT_TOKENS")
	if env == "" {
		return nil
	}

	for _, pair := range strings.Split(env, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			log.Printf("[bot_app] invalid BOT_TOKENS entry: %s", pair)
			continue
		}
		botID, token := parts[0], parts[1]

		bot, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Printf("[bot_app] init fail for %s: %v", botID, err)
			continue
		}

		app.bots[botID] = bot
		log.Printf("[bot_app] ready: @%s (%s)", bot.Self.UserName, botID)

		go app.runBotLoop(botID, bot)
	}

	return nil
}