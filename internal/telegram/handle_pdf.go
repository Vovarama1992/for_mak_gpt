package telegram

import (
	"bytes"
	"context"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handlePDF(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	mainKB tgbotapi.ReplyKeyboardMarkup,
) {
	chatID := msg.Chat.ID
	d := msg.Document

	log.Printf("[pdf] START bot=%s tg=%d filename=%s mime=%s",
		botID, tgID, d.FileName, d.MimeType)

	if !app.checkImageAllowed(ctx, botID, tgID) {
		bot.Send(tgbotapi.NewMessage(chatID, "📄 В этом тарифе разбор PDF недоступен."))
		return
	}

	// 1. TG FILE
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: d.FileID})
	if err != nil {
		log.Printf("[pdf] TG GetFile ERROR: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить PDF."))
		return
	}
	downloadURL := fileInfo.Link(bot.Token)
	log.Printf("[pdf] downloadURL=%s", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("[pdf] HTTP GET ERROR: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки PDF."))
		return
	}
	defer resp.Body.Close()

	// 2. PDF → IMAGES
	log.Printf("[pdf] converting via PDFService...")
	pages, err := app.PDFService.Convert(ctx, resp.Body)
	if err != nil {
		log.Printf("[pdf] CONVERT ERROR: %v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки PDF.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	log.Printf("[pdf] pages generated: %d", len(pages))

	// 3. SAVE EACH PAGE
	var firstImageURL *string

	for i, p := range pages {
		url, err := app.S3Service.SaveImage(
			ctx, botID, tgID,
			bytes.NewReader(p.Bytes),
			p.FileName,
			p.MimeType,
		)
		if err != nil {
			log.Printf("[pdf] S3 ERROR page=%d: %v", i+1, err)
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения страницы."))
			return
		}

		if firstImageURL == nil {
			firstImageURL = &url
		}

		app.RecordService.AddImage(ctx, botID, tgID, "user", url)
	}

	// 4. GPT CALL
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI читает PDF…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	reply, err := app.AiService.GetReply(
		ctx, botID, tgID,
		"image",
		"",            // пустой текст
		firstImageURL, // якорный файл
	)
	if err != nil {
		log.Printf("[pdf] GPT ERROR: %v", err)
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки PDF.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, reply))
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	log.Printf("[pdf] DONE bot=%s tg=%d", botID, tgID)
}
