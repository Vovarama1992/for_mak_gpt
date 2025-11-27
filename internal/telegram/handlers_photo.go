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
) {
	chatID := msg.Chat.ID

	photo := msg.Photo[len(msg.Photo)-1]
	log.Printf("[photo] start bot=%s tg=%d fileID=%s size=%dx%d",
		botID, tgID, photo.FileID, photo.Width, photo.Height)

	if !app.checkImageAllowed(ctx, botID, tgID) {
		bot.Send(tgbotapi.NewMessage(chatID,
			"🖼 В этом тарифе разбор по фото недоступен."))
		return
	}

	// === 1. Получаем файл у Telegram ===
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		log.Printf("[photo] get fail: %v", err)
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("❗ Ошибка получения фото\nБот: %s\nПользователь: %d\nFileID: %s",
				botID, tgID, photo.FileID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить фото."))
		return
	}

	downloadURL := fileInfo.Link(bot.Token)
	log.Printf("[photo] telegram_url=%s", downloadURL)

	// === 2. Скачиваем ===
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("[photo] download fail: %v", err)
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("❗ Ошибка загрузки фото\nБот: %s\nПользователь: %d\nURL: %s",
				botID, tgID, downloadURL))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки фото."))
		return
	}
	defer resp.Body.Close()

	filename := fmt.Sprintf("%s.jpg", photo.FileID)
	log.Printf("[photo] saving as %s", filename)

	// === 3. Сохраняем в S3 ===
	publicURL, err := app.S3Service.SaveImage(ctx, botID, tgID, resp.Body, filename, "image/jpeg")
	if err != nil {
		log.Printf("[photo] s3 save fail: %v", err)
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("❗ Ошибка сохранения в S3\nБот: %s\nПользователь: %d\nФайл: %s",
				botID, tgID, filename))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения фото."))
		return
	}
	log.Printf("[photo] s3_url=%s", publicURL)

	// === 4. История: user ===
	app.RecordService.AddText(ctx, botID, tgID, "user", publicURL)

	// === 5. Показываем индикатор ===
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	sentThinking, _ := bot.Send(thinking)

	// === 6. GPT ===
	gptInput := fmt.Sprintf("📷 Пользователь прислал изображение: %s", publicURL)
	reply, err := app.AiService.GetReply(ctx, botID, tgID, gptInput, &publicURL)

	if err != nil {
		log.Printf("[photo] ai fail: %v", err)
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("❗ Ошибка GPT\nБот: %s\nПользователь: %d\nФото: %s",
				botID, tgID, publicURL))

		// удаляем индикатор перед выходом
		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки фото."))
		return
	}

	log.Printf("[photo] ai_reply=%q", reply)

	// === 7. Отправляем ответ ===
	bot.Send(tgbotapi.NewMessage(chatID, reply))

	// === 8. История: tutor ===
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === 9. Удаляем индикатор в самом конце ===
	del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
	bot.Request(del)

	log.Printf("[photo] done botID=%s tgID=%d", botID, tgID)
}
