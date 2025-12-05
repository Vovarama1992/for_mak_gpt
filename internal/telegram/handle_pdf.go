package telegram

import (
	"bytes"
	"context"
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

	// тариф
	if !app.checkImageAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "📄 В этом тарифе разбор PDF недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	// 1. получить файл TG
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: d.FileID})
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить PDF."))
		return
	}
	downloadURL := fileInfo.Link(bot.Token)

	resp, err := http.Get(downloadURL)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки PDF."))
		return
	}
	defer resp.Body.Close()

	// 2. конвертация в картинки
	pages, err := app.PDFService.Convert(ctx, resp.Body)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки PDF."))
		return
	}

	// 3. сохраняем страницы в S3 + пишем историю
	var lastImageURL *string
	for _, p := range pages {
		url, err := app.S3Service.SaveImage(
			ctx, botID, tgID,
			bytes.NewReader(p.Bytes),
			p.FileName,
			p.MimeType,
		)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения изображения."))
			return
		}

		app.RecordService.AddImage(ctx, botID, tgID, "user", url)
		lastImageURL = &url // последняя страница для GPT
	}

	// 4. индикатор
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI читает PDF…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// 5. GPT (как фото)
	reply, err := app.AiService.GetReply(
		ctx, botID, tgID,
		"image", // ветка фото
		"Пользователь прислал PDF-файл.", // текстовая часть
		lastImageURL, // последняя страница как image_url
	)
	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки PDF."))
		return
	}

	// 6. ответ пользователю
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
}
