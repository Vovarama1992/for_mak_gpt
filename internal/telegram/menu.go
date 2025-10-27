package telegram

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Menu struct{}

func NewMenu() *Menu {
	return &Menu{}
}

func (m *Menu) ShowTariffs(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message, tariffSrv ports.TariffService) {
	tariffs, err := tariffSrv.ListAll(ctx)
	if err != nil || len(tariffs) == 0 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не удалось получить список тарифов"))
		return
	}

	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		apiBase = "https://aifull.ru/api"
	}

	text := "🚫 У вас нет активной подписки.\n\nВыберите подходящий тариф:"
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tariffs {
		q := url.Values{}
		q.Set("bot_id", bot.Self.UserName)
		q.Set("telegram_id", fmt.Sprintf("%d", msg.Chat.ID))
		q.Set("plan", t.Code)

		link := fmt.Sprintf("%s/subscribe/create?%s", apiBase, q.Encode())
		btn := tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf("%s — %.2f ₽", t.Name, t.Price), link)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msgCfg := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgCfg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	bot.Send(msgCfg)
}
