package telegram

import (
	"context"
	"log"

	"github.com/Vovarama1992/make_ziper/internal/ai"
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/speech"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotApp struct {
	SubscriptionService ports.SubscriptionService
	TariffService       ports.TariffService
	AiService           ai.Service
	SpeechService       *speech.Service
	RecordService       ports.RecordService
	S3Service           ports.S3Service

	BotsService bots.Service // ← добавлено
	bots        map[string]*tgbotapi.BotAPI
}

func (app *BotApp) InitBots() error {
	app.bots = make(map[string]*tgbotapi.BotAPI)

	// читаем все боты из БД
	cfgs, err := app.BotsService.ListAll(context.Background())
	if err != nil {
		return err
	}

	for _, cfg := range cfgs {
		if cfg.Token == "" {
			continue
		}

		bot, err := tgbotapi.NewBotAPI(cfg.Token)
		if err != nil {
			log.Printf("[bot_app] init fail for %s: %v", cfg.BotID, err)
			continue
		}

		app.bots[cfg.BotID] = bot
		log.Printf("[bot_app] ready: @%s (%s)", bot.Self.UserName, cfg.BotID)

		go app.runBotLoop(cfg.BotID, bot)
	}

	return nil
}
