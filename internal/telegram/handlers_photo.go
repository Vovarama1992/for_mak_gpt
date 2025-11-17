package telegram

import (
	"context"
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handlePhoto(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message, tgID int64) {

	chatID := msg.Chat.ID
	photo := msg.Photo[len(msg.Photo)-1]
	log.Printf("[photo] start bot=%s tg=%d fileID=%s size=%dx%d", botID, tgID, photo.FileID, photo.Width, photo.Height)

	if !app.checkImageAllowed(ctx, botID, tgID) {
		bot.Send(tgbotapi.NewMessage(chatID, "🖼 В этом тарифе разбор по фото недоступен."))
		return
	}

	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		log.Printf("[photo] get fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить фото."))
		return
	}

	downloadURL := fileInfo.Link(bot.Token)
	log.Printf("[photo] telegram_url=%s", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("[photo] download fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки фото."))
		return
	}
	defer resp.Body.Close()

	filename := fmt.Sprintf("%s.jpg", photo.FileID)
	log.Printf("[photo] saving as %s", filename)

	publicURL, err := app.S3Service.SaveImage(ctx, botID, tgID, resp.Body, filename, "image/jpeg")
	if err != nil {
		log.Printf("[photo] s3 save fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка хранения фото."))
		return
	}
	log.Printf("[photo] s3_url=%s", publicURL)

	_, err = app.RecordService.AddText(ctx, botID, tgID, "user", publicURL)
	if err != nil {
		log.Printf("[photo] AddImage record fail: %v", err)
	}

	gptInput := fmt.Sprintf("📷 Пользователь прислал изображение: %s", publicURL)
	reply, err := app.AiService.GetReply(ctx, botID, tgID, gptInput, &publicURL)
	if err != nil {
		log.Printf("[photo] ai fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки фото."))
		return
	}

	log.Printf("[photo] ai_reply=%q", reply)
	bot.Send(tgbotapi.NewMessage(chatID, reply))

	_, err = app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)
	if err != nil {
		log.Printf("[photo] AddText tutor fail: %v", err)
	}
}
