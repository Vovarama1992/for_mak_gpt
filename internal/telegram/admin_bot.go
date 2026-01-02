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
		log.Println("[admin-bot] token is empty, bot disabled")
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

	log.Println("[admin-bot] polling started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[admin-bot] context cancelled, stopping")
			return

		case upd := <-updates:
			if upd.Message != nil {
				log.Printf(
					"[admin-bot] incoming message from=%d text=%q",
					upd.Message.From.ID,
					upd.Message.Text,
				)

				a.handleMessage(upd.Message)
			}
		}
	}
}

// ==================================================
// SEND FROM USER → ADMIN
// ==================================================

func (a *AdminBot) Send(userID int64, text string) {
	admins := []int64{
		1139929360,
		6789440333,
	}

	for _, adminID := range admins {
		log.Printf(
			"[admin-bot] forward user=%d → admin=%d",
			userID,
			adminID,
		)

		_, err := a.bot.Send(tgbotapi.NewMessage(adminID, text))
		if err != nil {
			log.Printf(
				"[admin-bot] failed to send to admin=%d err=%v",
				adminID,
				err,
			)
			continue
		}

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

	log.Printf(
		"[admin-bot] admin message from=%d text=%q",
		adminID,
		msg.Text,
	)

	if msg.Text == "/start" {
		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"👋 Напиши сообщение — я передам его пользователю.",
		))
		return
	}

	ctxHelp, ok := a.app.adminHelpMode[adminID]
	if !ok {
		log.Printf(
			"[admin-bot] no active dialog for admin=%d",
			adminID,
		)

		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"❗ Нет активного диалога.",
		))
		return
	}

	reply := "💬 Ответ поддержки:\n\n" + msg.Text

	log.Printf(
		"[admin-bot] send reply admin=%d → user=%d",
		adminID,
		ctxHelp.UserID,
	)

	_, err := a.bot.Send(tgbotapi.NewMessage(
		ctxHelp.UserID,
		reply,
	))
	if err != nil {
		log.Printf(
			"[admin-bot] failed to send reply to user=%d err=%v",
			ctxHelp.UserID,
			err,
		)
		return
	}

	delete(a.app.adminHelpMode, adminID)

	a.bot.Send(tgbotapi.NewMessage(
		msg.Chat.ID,
		"✅ Ответ отправлен пользователю.",
	))
}
