package telegram

import (
	"context"
	"io"
	"log"
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

	log.Printf("[doc] start bot=%s tg=%d file=%s", botID, tgID, doc.FileName)

	// === 1. СКАЧИВАЕМ ===
	log.Printf("[doc] GetFile fileID=%s", doc.FileID)
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		log.Printf("[doc] GetFile ERROR: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить документ."))
		return
	}

	url := fileInfo.Link(bot.Token)
	log.Printf("[doc] downloading from=%s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[doc] download ERROR: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки документа."))
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("[doc] download BAD_STATUS=%d", resp.StatusCode)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки документа."))
		return
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[doc] read ERROR: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка чтения документа."))
		return
	}

	log.Printf("[doc] downloaded bytes=%d", len(raw))

	// === 2. КОНВЕРТИРУЕМ DOC → TEXT ===
	log.Printf("[doc] converting…")
	text, err := app.DocService.Convert(ctx, raw)
	log.Printf("[doc] convert result len=%d err=%v", len(text), err)

	if err != nil || len(text) == 0 {
		log.Printf("[doc] convert FAIL (text empty or error)")
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа."))
		return
	}

	// === ДАЛЬШЕ — ТОЧНО handleText ===

	// === 3. показываем 'AI думает…' ===
	log.Printf("[doc] show thinking")
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI читает документ…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// === 4. GPT ===
	log.Printf("[doc] GPT request…")
	reply, err := app.AiService.GetReply(
		ctx,
		botID,
		tgID,
		"text",
		text,
		nil,
	)
	log.Printf("[doc] GPT done err=%v", err)

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа AI."))
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		return
	}

	// === 5. отправляем ответ ===
	log.Printf("[doc] send reply len=%d", len(reply))
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	// === 6. сохраняем историю ===
	log.Printf("[doc] save history")
	app.RecordService.AddText(ctx, botID, tgID, "user", text)
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === 7. убираем индикатор ===
	//log.Printf("[doc] delete thinking msg=%d", sentThinking.MessageID)
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	log.Printf("[doc] done bot=%s tg=%d", botID, tgID)
}
