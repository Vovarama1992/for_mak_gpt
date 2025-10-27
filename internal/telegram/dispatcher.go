package telegram

import (
	"context"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Dispatcher struct {
	bot             *tgbotapi.BotAPI
	subscriptionSrv ports.SubscriptionService
	tariffSrv       ports.TariffService
	menu            *Menu
}

func NewDispatcher(
	bot *tgbotapi.BotAPI,
	subSrv ports.SubscriptionService,
	tariffSrv ports.TariffService,
) *Dispatcher {
	return &Dispatcher{
		bot:             bot,
		subscriptionSrv: subSrv,
		tariffSrv:       tariffSrv,
		menu:            NewMenu(),
	}
}

// CheckAndShowMenu проверяет подписку и показывает меню, если её нет
func (d *Dispatcher) CheckAndShowMenu(ctx context.Context, telegramID int64) {
	status, err := d.subscriptionSrv.GetStatus(ctx, d.bot.Self.UserName, telegramID)
	if err != nil {
		return
	}

	if status == "active" || status == "pending" {
		return // подписка есть, меню не показываем
	}

	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: telegramID}}
	d.menu.ShowTariffs(ctx, d.bot, msg, d.tariffSrv)
}
