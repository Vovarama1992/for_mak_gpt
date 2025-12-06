package telegram

import (
	"bytes"
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

	// 1. Получаем файл из Telegram
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить документ."))
		return
	}

	downloadURL := fileInfo.Link(bot.Token)

	resp, err := http.Get(downloadURL)
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

	// 2. DOC → PDF → JPEG
	pages, err := app.DocService.Convert(ctx, raw)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа."))
		return
	}

	// 3. Сохраняем ВСЕ страницы в S3 + историю (как в PDF-ветке)
	for _, p := range pages {
		url, err := app.S3Service.SaveImage(
			ctx, botID, tgID,
			bytes.NewReader(p.Bytes),
			p.FileName,
			p.MimeType,
		)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка сохранения изображения."))
			return
		}

		app.RecordService.AddImage(ctx, botID, tgID, "user", url)
	}

	// 4. Индикатор
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI читает документ…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// 5. GPT → ветка image (БЕЗ передачи lastImageURL — история уже содержит всё)
	reply, err := app.AiService.GetReply(
		ctx, botID, tgID,
		"image",
		" ",
		nil,
	)
	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки документа AI."))
		return
	}

	// 6. Ответ
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	// 7. Удаляем индикатор
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
}
