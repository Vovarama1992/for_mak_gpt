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

	photo := msg.Photo[len(msg.Photo)-1] // лучшее качество
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

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка получения фото\n\nБот: %s\nПользователь: %d\nFileID: %s\n\nЧто проверить:\n— токен Telegram\n— доступность файла у Telegram",
				botID, tgID, photo.FileID,
			),
		)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить фото."))
		return
	}

	downloadURL := fileInfo.Link(bot.Token)
	log.Printf("[photo] telegram_url=%s", downloadURL)

	// === 2. Скачиваем фото ===
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("[photo] download fail: %v", err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка загрузки фото c серверов Telegram\n\nБот: %s\nПользователь: %d\nURL: %s\n\nЧто проверить:\n— интернет бота\n— корректность FileID\n— актуальность токена",
				botID, tgID, downloadURL,
			),
		)

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

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка сохранения фото в S3\n\nБот: %s\nПользователь: %d\nФайл: %s\n\nЧто проверить:\n— S3 credentials\n— bucket\n— права записи\n— content-type",
				botID, tgID, filename,
			),
		)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения фото."))
		return
	}
	log.Printf("[photo] s3_url=%s", publicURL)

	// === 4. Записываем в историю (user) ===
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "user", publicURL); err != nil {
		log.Printf("[photo] AddImage record fail: %v", err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка записи фото в историю (user)\n\nБот: %s\nПользователь: %d\nURL: %s\n\nЧто проверить:\n— таблицу records\n— соединение с БД",
				botID, tgID, publicURL,
			),
		)
	}

	// === 💭 5. Показываем "думаю..." ===
	thinkingMsg := tgbotapi.NewMessage(chatID, "💭 Думаю...")
	sentThinking, _ := bot.Send(thinkingMsg)

	// === 🤖 6. GPT запрос ===
	gptInput := fmt.Sprintf("📷 Пользователь прислал изображение: %s", publicURL)
	reply, err := app.AiService.GetReply(ctx, botID, tgID, gptInput, &publicURL)

	// === ❌ удаляем сообщение "думаю..." ===
	delReq := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
	bot.Request(delReq)

	if err != nil {
		log.Printf("[photo] ai fail: %v", err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка GPT при разборе фото\n\nБот: %s\nПользователь: %d\nФото: %s\n\nЧто проверить:\n— модель GPT\n— токен OpenAI\n— корректность построения input",
				botID, tgID, publicURL,
			),
		)

		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки фото."))
		return
	}

	log.Printf("[photo] ai_reply=%q", reply)

	// === 7. Отправляем ответ пользователю ===
	if _, err := bot.Send(tgbotapi.NewMessage(chatID, reply)); err != nil {
		log.Printf("[photo] send reply fail: %v", err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка отправки ответа пользователю\n\nБот: %s\nПользователь: %d\nОтвет: %s\n\nЧто проверить:\n— токен Telegram\n— лимиты на отправку сообщений",
				botID, tgID, reply,
			),
		)
		return
	}

	// === 8. Записываем в историю (tutor) ===
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "tutor", reply); err != nil {
		log.Printf("[photo] AddText tutor fail: %v", err)

		app.ErrorNotify.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf(
				"❗ Ошибка записи истории (tutor)\n\nБот: %s\nПользователь: %d\nОтвет: %s\n\nЧто проверить:\n— таблицу records\n— соединение с БД",
				botID, tgID, reply,
			),
		)
	}

	log.Printf("[photo] done botID=%s tgID=%d", botID, tgID)
}
