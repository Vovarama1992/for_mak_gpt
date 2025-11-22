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

	// === 1. GPT ответ ===
	reply, err := app.AiService.GetReply(ctx, botID, tgID, userText, nil)
	if err != nil {
		log.Printf("[text] ai reply fail botID=%s tgID=%d: %v", botID, tgID, err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка GPT ответа\n\nБот: %s\nПользователь: %d\nТекст: %q\n\nЧто проверить:\n— модель в настройках бота\n— токен OpenAI\n— историю диалога",
				botID, tgID, userText,
			),
		)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке запроса."))
		return
	}

	// === 2. Отправляем ответ пользователю ===
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, reply)); err != nil {
		log.Printf("[text] bot.Send reply fail botID=%s tgID=%d: %v", botID, tgID, err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Telegram send error\n\nБот: %s\nПользователь: %d\nОтвет: %q\n\nЧто проверить:\n— токен Telegram\n— лимиты отправки\n— формат сообщения",
				botID, tgID, reply,
			),
		)
		return
	}

	// === 3. Пишем историю (user → tutor) ===

	// user message
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "user", userText); err != nil {
		log.Printf("[text] AddText user fail botID=%s tgID=%d: %v", botID, tgID, err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка записи истории (user)\n\nБот: %s\nПользователь: %d\nТекст: %q\n\nЧто проверить:\n— таблицу records\n— строку подключения к БД",
				botID, tgID, userText,
			),
		)
	}

	// gpt reply
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "tutor", reply); err != nil {
		log.Printf("[text] AddText tutor fail botID=%s tgID=%d: %v", botID, tgID, err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка записи истории (tutor)\n\nБот: %s\nПользователь: %d\nОтвет: %q\n\nЧто проверить:\n— таблицу records\n— строку подключения к БД",
				botID, tgID, reply,
			),
		)
	}

	log.Printf("[text] done botID=%s tgID=%d", botID, tgID)
}
