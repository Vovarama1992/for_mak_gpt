package telegram

import (
	"context"
	"fmt"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Menu struct{}

func NewMenu() *Menu { return &Menu{} }

// ShowTariffs — показывает пользователю список тарифов (pending создаётся в RunListeners)
func (m *Menu) ShowTariffs(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tariffSrv ports.TariffService,
) {
	tariffs, err := tariffSrv.ListAll(ctx)
	if err != nil || len(tariffs) == 0 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не удалось получить список тарифов"))
		return
	}

	text := "🚫 У вас нет активной подписки.\n\nВыберите подходящий тариф:"
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, t := range tariffs {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s — %.2f ₽", t.Name, t.Price),
			fmt.Sprintf("subscribe:%s", t.Code),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msgCfg := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := bot.Send(msgCfg); err != nil {
		fmt.Printf("[ShowTariffs] send error: %v\n", err)
	}
}
