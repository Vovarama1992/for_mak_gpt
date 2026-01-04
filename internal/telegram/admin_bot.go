package telegram

import (
	"context"
	"log"
	"regexp"
	"strconv"

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
					"[admin-bot] received message from user=%d chat=%d text=%q",
					upd.Message.From.ID,
					upd.Message.Chat.ID,
					upd.Message.Text,
				)
				a.handleMessage(upd.Message)
			}
		}
	}
}

// ==================================================
// USER → ADMIN (FORWARD)
// ==================================================

func (a *AdminBot) Send(userID int64, text string) {
	admins := []int64{
		1139929360,
		6789440333,
	}

	for _, adminID := range admins {
		log.Printf(
			"[admin-bot] forwarding message user=%d → admin=%d",
			userID,
			adminID,
		)

		msg := tgbotapi.NewMessage(adminID, text)
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

		if _, err := a.bot.Send(msg); err != nil {
			log.Printf(
				"[admin-bot] failed to forward user=%d → admin=%d err=%v",
				userID,
				adminID,
				err,
			)
		}
	}
}

// ==================================================
// ADMIN MESSAGE HANDLER
// ==================================================

func (a *AdminBot) handleMessage(msg *tgbotapi.Message) {
	log.Printf(
		"[admin-bot] admin message from=%d text=%q reply=%v",
		msg.From.ID,
		msg.Text,
		msg.ReplyToMessage != nil,
	)

	if msg.Text == "/start" {
		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"👋 Это бот поддержки. Отвечай reply на сообщения пользователей.",
		))
		return
	}

	// принимаем ТОЛЬКО ответы reply
	if msg.ReplyToMessage == nil {
		log.Printf(
			"[admin-bot] ignore admin=%d message without reply",
			msg.From.ID,
		)

		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"❗ Ответь reply на сообщение пользователя.",
		))
		return
	}

	userID, ok := extractUserID(msg.ReplyToMessage.Text)
	if !ok {
		log.Printf(
			"[admin-bot] failed to extract userID from reply text=%q",
			msg.ReplyToMessage.Text,
		)

		a.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"❗ Не удалось определить пользователя.",
		))
		return
	}

	log.Printf(
		"[admin-bot] sending reply admin=%d → user=%d",
		msg.From.ID,
		userID,
	)

	reply := "💬 Ответ поддержки:\n\n" + msg.Text

	if _, err := a.bot.Send(tgbotapi.NewMessage(userID, reply)); err != nil {
		log.Printf(
			"[admin-bot] failed to send reply admin=%d → user=%d err=%v",
			msg.From.ID,
			userID,
			err,
		)
		return
	}

	a.bot.Send(tgbotapi.NewMessage(
		msg.Chat.ID,
		"✅ Ответ отправлен пользователю.",
	))
}

// ==================================================
// UTILS
// ==================================================

func extractUserID(text string) (int64, bool) {
	re := regexp.MustCompile(`UserID:\s*(\d+)`)
	m := re.FindStringSubmatch(text)
	if len(m) != 2 {
		return 0, false
	}

	id, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
