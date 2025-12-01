package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Vovarama1992/make_ziper/internal/speech"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleVoice(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message, tgID int64) {

	chatID := msg.Chat.ID
	fileID := msg.Voice.FileID

	log.Printf("[voice] start botID=%s tgID=%d fileID=%s", botID, tgID, fileID)

	if !app.checkVoiceAllowed(ctx, botID, tgID) {
		bot.Send(tgbotapi.NewMessage(chatID, "🔇 В этом тарифе голос временно недоступен."))
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
			fmt.Sprintf("Не удалось получить файл: tg=%d fileID=%s", tgID, fileID))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое."))
		return
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка загрузки файла: %s", url))
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
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка сохранения голосового."))
		return
	}
	out.Close()
	defer os.Remove(path)

	text, err := app.SpeechService.Transcribe(ctx, botID, path)
	if err != nil {
		app.ErrorNotify.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка распознавания речи: %s", path))
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос."))
		return
	}

	app.RecordService.AddText(ctx, botID, tgID, "user", text)

	// === показываем индикатор ===
	thinkingMsg := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	sentThinking, _ := bot.Send(thinkingMsg)

	// === GPT ===
	reply, err := app.AiService.GetReply(
		ctx,
		botID,
		tgID,
		"voice",
		text,
		nil,
	)

	// === синтез ответа ===
	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {

		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)

		app.ErrorNotify.Notify(ctx, botID, err,
			"Ошибка синтеза ответа в аудио")
		bot.Send(tgbotapi.NewMessage(chatID, reply))
		return
	}
	defer os.Remove(outVoice)

	// === списываем минуты за TTS ===
	if durSec, err := speech.AudioDuration(outVoice); err == nil {
		usedReplyMinutes := durSec / 60.0
		go func() {
			ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedReplyMinutes)
			if err != nil {
				app.ErrorNotify.Notify(ctx, botID, err,
					fmt.Sprintf("Ошибка списания TTS минут: tg=%d", tgID))
				return
			}
			if !ok {
				log.Printf("[voice] async: no voice minutes left for TTS tgID=%d", tgID)
			}
		}()
	}

	// === отправляем аудио ===
	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	bot.Send(voice)

	// === пишем историю ===
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// === Удаляем индикатор в самом конце ===
	del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
	bot.Request(del)
}
