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
	mainKB tgbotapi.ReplyKeyboardMarkup, // ← ДОБАВИЛИ
) {
	chatID := msg.Chat.ID

	photo := msg.Photo[len(msg.Photo)-1]
	log.Printf("[photo] start bot=%s tg=%d fileID=%s size=%dx%d",
		botID, tgID, photo.FileID, photo.Width, photo.Height)

	// тариф не позволяет
	if !app.checkImageAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "🖼 В этом тарифе разбор по фото недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	// === 1. Получаем файл TG ===
	fileInfo, err := bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		bot.Send(withKB(chatID, "⚠️ Не удалось получить фото.", mainKB))
		return
	}

	downloadURL := fileInfo.Link(bot.Token)

	// === 2. Скачиваем ===
	resp, err := http.Get(downloadURL)
	if err != nil {
		bot.Send(withKB(chatID, "⚠️ Ошибка загрузки фото.", mainKB))
		return
	}
	defer resp.Body.Close()

	filename := fmt.Sprintf("%s.jpg", photo.FileID)

	// === 3. Сохраняем в S3 ===
	publicURL, err := app.S3Service.SaveImage(ctx, botID, tgID, resp.Body, filename, "image/jpeg")
	if err != nil {
		bot.Send(withKB(chatID, "⚠️ Ошибка хранения фото.", mainKB))
		return
	}

	// === 4. История ===
	app.RecordService.AddImage(ctx, botID, tgID, "user", publicURL)

	// === 5. Индикатор «думает» ===
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// === 6. GPT ===
	gptInput := "📷 Пользователь прислал изображение."
	reply, err := app.AiService.GetReply(ctx, botID, tgID, "text", gptInput, &publicURL)
	if err != nil {
		// убрать индикатор
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		bot.Send(withKB(chatID, "⚠️ Ошибка обработки фото.", mainKB))
		return
	}

	// === 7. Ответ GPT (С ОБЯЗАТЕЛЬНОЙ клавиатурой) ===
	out := tgbotapi.NewMessage(chatID, reply)
	out.ReplyMarkup = mainKB
	bot.Send(out)

	// === 8. История ===
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === 9. Удаляем индикатор ===
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	log.Printf("[photo] done botID=%s tgID=%d", botID, tgID)
}

// вспомогательная функция
func withKB(chatID int64, text string, kb tgbotapi.ReplyKeyboardMarkup) tgbotapi.MessageConfig {
	m := tgbotapi.NewMessage(chatID, text)
	m.ReplyMarkup = kb
	return m
}
