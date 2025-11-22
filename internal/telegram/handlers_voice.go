package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleVoice(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message, tgID int64) {

	chatID := msg.Chat.ID
	fileID := msg.Voice.FileID

	log.Printf("[voice] start botID=%s tgID=%d fileID=%s", botID, tgID, fileID)

	if !app.checkVoiceAllowed(ctx, botID, tgID) {
		bot.Send(tgbotapi.NewMessage(chatID, "🔇 В этом тарифе голос временно недоступен."))
		log.Printf("[voice] not allowed botID=%s tgID=%d", botID, tgID)
		return
	}

	usedMinutes := float64(msg.Voice.Duration) / 60.0

	go func() {
		ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedMinutes)
		if err != nil {
			app.ErrorNotify.Notify(ctx, botID, err,
				fmt.Sprintf("Ошибка списания голосовых минут: tg=%d", tgID))
			return
		}
		if !ok {
			log.Printf("[voice] async: no voice minutes left for tgID=%d", tgID)
		}
	}()

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Не удалось получить файл голосового: tg=%d fileID=%s", tgID, fileID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое."))
		return
	}

	url := file.Link(bot.Token)
	log.Printf("[voice] downloading from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка загрузки файла по ссылке: %s", url))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при загрузке голосового."))
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, err := os.Create(path)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка создания временного файла: %s", path))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке голосового."))
		return
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка записи файла: %s", path))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при сохранении голосового."))
		return
	}
	out.Close()
	defer os.Remove(path)

	log.Printf("[voice] saved to %s", path)

	text, err := app.SpeechService.Transcribe(ctx, botID, path)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка распознавания речи: файл %s", path))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос."))
		return
	}

	if _, err := app.RecordService.AddText(ctx, botID, tgID, "user", text); err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			"Ошибка записи текста пользователя в историю диалога")
	}

	reply, err := app.AiService.GetReply(ctx, botID, tgID, text, nil)
	if err != nil {
		// AiService сам шлёт уведомление
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при ответе."))
		return
	}

	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			"Ошибка синтеза ответа в аудио (voice_id неверный?)")
		bot.Send(tgbotapi.NewMessage(chatID, reply))
		return
	}
	defer os.Remove(outVoice)

	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	if _, err := bot.Send(voice); err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			"Ошибка отправки голосового ответа пользователю")
	}

	if _, err := app.RecordService.AddText(ctx, botID, tgID, "tutor", reply); err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			"Ошибка записи ответа GPT в историю диалога")
	}
}
