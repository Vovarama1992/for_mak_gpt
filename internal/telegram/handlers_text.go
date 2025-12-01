package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleText(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
) {
	chatID := msg.Chat.ID
	userText := msg.Text

	log.Printf("[text] start botID=%s tgID=%d", botID, tgID)

	// === 0. показываем 'AI думает…' ===
	thinkingMsg := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	sentThinking, _ := bot.Send(thinkingMsg)

	// === 1. GPT ===
	reply, err := app.AiService.GetReply(
		ctx,
		botID,
		tgID,
		"text",
		userText,
		nil,
	)

	if err != nil {
		log.Printf("[text] ai reply fail botID=%s tgID=%d: %v", botID, tgID, err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf("❗ Ошибка GPT ответа\n\nБот: %s\nПользователь: %d\nТекст: %q",
				botID, tgID, userText),
		)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке запроса."))

		// === удаляем индикатор ===
		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)
		return
	}

	// === 2. отправляем ответ ===
	bot.Send(tgbotapi.NewMessage(chatID, reply))

	// === 3. пишем историю ===
	app.RecordService.AddText(ctx, botID, tgID, "user", userText)
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === 4. удаляем 'AI думает…' ===
	del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
	bot.Request(del)

	log.Printf("[text] done botID=%s tgID=%d", botID, tgID)
}
