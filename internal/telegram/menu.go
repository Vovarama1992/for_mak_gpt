package telegram

import (
	"context"
	"fmt"

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

	text := "🚫 У вас нет активной подписки.\n\nВыберите подходящий тариф:"
	list := ""
	for _, t := range tariffs {
		list += fmt.Sprintf("• %s — %.2f ₽ / %d дн.\n", t.Name, t.Price, t.PeriodDays)
	}

	msgCfg := tgbotapi.NewMessage(msg.Chat.ID, text+"\n\n"+list)
	msgCfg.ReplyMarkup = TariffsKeyboard(tariffs)
	bot.Send(msgCfg)
}

func (m *Menu) HandleCallback(ctx context.Context, bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, subSrv ports.SubscriptionService) {
	code := cb.Data
	if code == "" {
		bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Некорректный тариф"))
		bot.Request(tgbotapi.NewCallback(cb.ID, ""))
		return
	}

	subSrv.Create(ctx, bot.Self.UserName, cb.Message.Chat.ID, code)
	text := fmt.Sprintf("✅ Подписка '%s' оформлена!\nОжидайте активации.", code)
	bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, text))

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}
