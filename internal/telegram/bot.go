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
}

func (app *BotApp) RunAll() error {
	tokensEnv := os.Getenv("BOT_TOKENS")
	if tokensEnv == "" {
		log.Println("no BOT_TOKENS provided")
		return nil
	}

	tokens := strings.Split(tokensEnv, ",")
	var wg sync.WaitGroup

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		wg.Add(1)
		go func(tok string) {
			defer wg.Done()

			bot, err := tgbotapi.NewBotAPI(tok)
			if err != nil {
				log.Printf("failed to start bot: %v", err)
				return
			}
			log.Printf("bot started: %s", bot.Self.UserName)

			dispatcher := NewDispatcher(
				bot,
				app.SubscriptionService,
				app.TariffService,
			)
			dispatcher.Run()
		}(token)
	}

	wg.Wait()
	return nil
}
