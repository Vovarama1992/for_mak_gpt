package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminBot struct {
	bot *tgbotapi.BotAPI
	app *BotApp
}

// ==================================================
// INIT
// ==================================================

func (app *BotApp) InitAdminBot(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	log.Printf("[admin-bot] started: @%s", bot.Self.UserName)

	admin := &AdminBot{
		bot: bot,
		app: app,
	}

	app.adminBot = admin

	go admin.run(ctx)

	return nil
}

// ==================================================
// MAIN LOOP
// ==================================================

func (a *AdminBot) run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return

		case upd := <-updates:
			if upd.Message != nil {
				a.handleMessage(upd.Message)
			}
		}
	}
}

// ==================================================
// SEND FROM USER-BOT → ADMIN
// ==================================================

func (a *AdminBot) Send(userID int64, text string) {
	admins := []int64{
		1139929360,
		6789440333,
	}

	for _, adminID := range admins {
		a.bot.Send(tgbotapi.NewMessage(adminID, text))

		a.app.adminHelpMode[adminID] = &AdminHelpContext{
			UserID: userID,
		}
	}
}

// ==================================================
// HANDLE ADMIN MESSAGE
// ==================================================

func (a *AdminBot) handleMessage(msg *tgbotapi.Message) {
	adminID := msg.From.ID

	// старт диалога
	if msg.Text == "/start" {
		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"👋 Напиши сообщение — я передам его пользователю.",
		))
		return
	}

	ctxHelp, ok := a.app.adminHelpMode[adminID]
	if !ok {
		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"❗ Нет активного диалога.",
		))
		return
	}

	reply := "💬 Ответ поддержки:\n\n" + msg.Text

	// ⬇️ напрямую пользователю
	a.bot.Send(tgbotapi.NewMessage(
		ctxHelp.UserID,
		reply,
	))

	delete(a.app.adminHelpMode, adminID)

	a.bot.Send(tgbotapi.NewMessage(
		msg.Chat.ID,
		"✅ Ответ отправлен пользователю.",
	))
}
