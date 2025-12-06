package telegram

import (
	"context"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleDoc(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	mainKB tgbotapi.ReplyKeyboardMarkup,
) {
	chatID := msg.Chat.ID
	doc := msg.Document

	// тариф
	if !app.checkImageAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "📄 В этом тарифе разбор документов недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	// ==== 1. скачиваем файл ====
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить документ."))
		return
	}

	resp, err := http.Get(fileInfo.Link(bot.Token))
	if err != nil || resp.StatusCode != 200 {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки документа."))
		return
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка чтения документа."))
		return
	}

	// ==== 2. DOC → TEXT ====
	text, err := app.DocService.Convert(ctx, raw) // ← ровно string, как у тебя в интерфейсе
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа."))
		return
	}

	if len(text) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось извлечь текст из документа."))
		return
	}

	// ==== 3. сохраняем текст пользователя ====
	app.RecordService.AddText(ctx, botID, tgID, "user", text)

	// ==== 4. индикатор ====
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI читает документ…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// ==== 5. GPT как текст ====
	reply, err := app.AiService.GetReply(
		ctx,
		botID,
		tgID,
		"text", // ← ОЧЕНЬ ВАЖНО: это текстовая ветка
		text,
		nil,
	)
	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа AI."))
		return
	}

	// ==== 6. отправляем ответ ====
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	// ==== 7. сохраняем ответ в историю ====
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// ==== 8. убираем индикатор ====
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
}
