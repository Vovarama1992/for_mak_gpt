package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleText(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	mainKB tgbotapi.ReplyKeyboardMarkup, // ← добавили
) {
	chatID := msg.Chat.ID
	userText := msg.Text

	log.Printf("[text] start botID=%s tgID=%d", botID, tgID)

	// === 0. показываем 'AI думает…' ===
	thinkingMsg := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinkingMsg.ReplyMarkup = mainKB // ← держим меню
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

		out := tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке запроса.")
		out.ReplyMarkup = mainKB // ← держим меню
		bot.Send(out)

		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)
		return
	}

	// === 2. GPT ответ ===
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB // ← КРИТИЧЕСКОЕ МЕСТО
	bot.Send(out)

	// === 3. история ===
	app.RecordService.AddText(ctx, botID, tgID, "user", userText)
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === 4. удаляем индикатор "думает" ===
	del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
	bot.Request(del)

	log.Printf("[text] done botID=%s tgID=%d", botID, tgID)
}
