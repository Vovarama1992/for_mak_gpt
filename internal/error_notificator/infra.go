package error_notificator

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const adminChatID int64 = 1139929360

type Infra struct {
	bots map[string]*tgbotapi.BotAPI
}

func NewInfra(bots map[string]*tgbotapi.BotAPI) *Infra {
	return &Infra{bots: bots}
}

// SetBots — позволяет передать карту ботов ПОСЛЕ того, как они инициализировались
func (i *Infra) SetBots(bots map[string]*tgbotapi.BotAPI) {
	i.bots = bots
}

func (i *Infra) Notify(ctx context.Context, botID string, err error, details string) error {
	bot, ok := i.bots[botID]
	if !ok || bot == nil {
		log.Printf("[error_notificator] botID=%s not found", botID)
		return fmt.Errorf("bot not found: %s", botID)
	}

	text := fmt.Sprintf(
		"❗ Ошибка в боте (%s)\n\nОшибка: %v\n\nДетали: %s",
		botID,
		err,
		details,
	)

	msg := tgbotapi.NewMessage(adminChatID, text)

	_, sendErr := bot.Send(msg)
	if sendErr != nil {
		log.Printf("[error_notificator] send fail: %v", sendErr)
		return sendErr
	}

	return nil
}
