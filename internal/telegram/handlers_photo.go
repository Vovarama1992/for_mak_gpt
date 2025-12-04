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
	// ОПРЕДЕЛЯЕМ ФАЙЛ: фото или документ
	//--------------------------------------------------------
	var fileID, filename, contentType string

	if msg.Document != nil {
		// документ → как фото
		d := msg.Document
		fileID = d.FileID
		filename = d.FileName
		contentType = d.MimeType

		log.Printf("[document->photo] bot=%s tg=%d file=%s mime=%s",
			botID, tgID, fileID, contentType)

	} else {
		// обычное фото
		photo := msg.Photo[len(msg.Photo)-1]
		fileID = photo.FileID
		filename = fmt.Sprintf("%s.jpg", photo.FileID)
		contentType = "image/jpeg"

		log.Printf("[photo] bot=%s tg=%d file=%s size=%dx%d",
			botID, tgID, photo.FileID, photo.Width, photo.Height)
	}

	//--------------------------------------------------------
	// тариф
	//--------------------------------------------------------
	if !app.checkImageAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "🖼 В этом тарифе разбор изображения недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	//--------------------------------------------------------
	// 1. Получаем файл из Telegram
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
	// 4. История (как фото)
	//--------------------------------------------------------
	app.RecordService.AddImage(ctx, botID, tgID, "user", publicURL)

	//--------------------------------------------------------
	// 5. Индикатор «думает»
	//--------------------------------------------------------
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	//--------------------------------------------------------
	// 6. GPT
	//--------------------------------------------------------
	gptInput := "📷 Пользователь прислал изображение."

	reply, err := app.AiService.GetReply(
		ctx, botID, tgID,
		"photo",
		gptInput,
		&publicURL,
	)

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
