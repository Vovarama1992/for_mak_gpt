package notificator

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	adminChatID1 int64 = 1139929360
	adminChatID2 int64 = 6789440333
)

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

	admins := []int64{adminChatID1, adminChatID2}

	for _, chatID := range admins {
		_, sendErr := bot.Send(tgbotapi.NewMessage(chatID, text))
		if sendErr != nil {
			log.Printf("[error_notificator] send fail to %d: %v", chatID, sendErr)
			return sendErr
		}
	}

	return nil
}

func (i *Infra) UserNotify(
	ctx context.Context,
	botID string,
	chatID int64,
	text string,
) error {
	bot, ok := i.bots[botID]
	if !ok || bot == nil {
		return fmt.Errorf("bot not found: %s", botID)
	}

	_, err := bot.Send(tgbotapi.NewMessage(chatID, text))
	return err
}
