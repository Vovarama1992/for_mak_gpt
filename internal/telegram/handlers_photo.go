package telegram

import (
	"context"
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handlePhoto(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	mainKB tgbotapi.ReplyKeyboardMarkup,
) {
	chatID := msg.Chat.ID

	//--------------------------------------------------------
	// ОПРЕДЕЛЯЕМ ФАЙЛ
	//--------------------------------------------------------
	var fileID, filename, contentType string
	var isDocument bool

	if msg.Document != nil {
		d := msg.Document
		isDocument = true
		fileID = d.FileID
		filename = d.FileName
		contentType = d.MimeType

		log.Printf("[document] bot=%s tg=%d file=%s mime=%s",
			botID, tgID, fileID, contentType)
	} else {
		p := msg.Photo[len(msg.Photo)-1]
		fileID = p.FileID
		filename = fmt.Sprintf("%s.jpg", p.FileID)
		contentType = "image/jpeg"

		log.Printf("[photo] bot=%s tg=%d file=%s size=%dx%d",
			botID, tgID, p.FileID, p.Width, p.Height)
	}

	//--------------------------------------------------------
	// тариф
	//--------------------------------------------------------
	if !app.checkImageAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "🖼 В этом тарифе разбор изображений недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	//--------------------------------------------------------
	// 1. Получаем файл
	//--------------------------------------------------------
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить файл.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	downloadURL := fileInfo.Link(bot.Token)

	//--------------------------------------------------------
	// 2. Скачиваем
	//--------------------------------------------------------
	resp, err := http.Get(downloadURL)
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки файла.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer resp.Body.Close()

	//--------------------------------------------------------
	// 3. Сохраняем в S3
	//--------------------------------------------------------
	publicURL, err := app.S3Service.SaveImage(ctx, botID, tgID, resp.Body, filename, contentType)
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения файла.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	//--------------------------------------------------------
	// 4. История
	//--------------------------------------------------------
	if isDocument {
		app.RecordService.AddText(ctx, botID, tgID, "user", "📄 Документ: "+publicURL)
	} else {
		app.RecordService.AddImage(ctx, botID, tgID, "user", publicURL)
	}

	//--------------------------------------------------------
	// 5. Индикатор «думает»
	//--------------------------------------------------------
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	//--------------------------------------------------------
	// 6. GPT
	//--------------------------------------------------------
	var reply string

	if isDocument {
		// документ → обычный текстовый URL
		gptInput := "📄 Пользователь прислал документ: " + publicURL

		reply, err = app.AiService.GetReply(
			ctx, botID, tgID,
			"text", // документ идёт как текст
			gptInput,
			nil, // без image_url
		)

	} else {
		// фото → как раньше
		gptInput := "📷 Пользователь прислал изображение."

		reply, err = app.AiService.GetReply(
			ctx, botID, tgID,
			"photo",
			gptInput,
			&publicURL, // image_url
		)
	}

	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки файла.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	//--------------------------------------------------------
	// 7. Ответ
	//--------------------------------------------------------
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	//--------------------------------------------------------
	// 8. История
	//--------------------------------------------------------
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	//--------------------------------------------------------
	// 9. Удаляем индикатор
	//--------------------------------------------------------
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	log.Printf("[photo/document] done botID=%s tgID=%d", botID, tgID)
}
