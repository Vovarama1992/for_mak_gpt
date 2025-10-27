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

func (d *Dispatcher) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := d.bot.GetUpdatesChan(u)

	for update := range updates {
		go d.routeUpdate(context.Background(), update)
	}
}

func (d *Dispatcher) routeUpdate(ctx context.Context, upd tgbotapi.Update) {
	switch {
	case upd.Message != nil:
		d.handleMessage(ctx, upd.Message)
	case upd.CallbackQuery != nil:
		d.handleCallback(ctx, upd.CallbackQuery)
	}
}

func (d *Dispatcher) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	status, err := d.subscriptionSrv.GetStatus(ctx, d.bot.Self.UserName, msg.Chat.ID)
	if err != nil || (status != "active" && status != "pending") {
		d.menu.ShowTariffs(ctx, d.bot, msg, d.tariffSrv)
		return
	}

	// здесь идёт логика обработки сообщений (GPT и т.п.)
}

func (d *Dispatcher) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	d.menu.HandleCallback(ctx, d.bot, cb, d.subscriptionSrv)
}
