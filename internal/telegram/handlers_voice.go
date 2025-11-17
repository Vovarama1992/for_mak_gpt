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

	// -----------------------
	// асинхронное списание голосовых минут
	// -----------------------
	usedMinutes := float64(msg.Voice.Duration) / 60.0

	go func() {
		ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedMinutes)
		if err != nil {
			log.Printf("[voice] async UseVoiceMinutes fail: %v", err)
			return
		}
		if !ok {
			log.Printf("[voice] async: no voice minutes left for tgID=%d", tgID)
		}
		log.Printf("[voice] async deducted %.2fmin ok", usedMinutes)
	}()

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("[voice] get file fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое."))
		return
	}

	url := file.Link(bot.Token)
	log.Printf("[voice] downloading from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[voice] download fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при загрузке голосового."))
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, err := os.Create(path)
	if err != nil {
		log.Printf("[voice] create tmp fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке голосового."))
		return
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("[voice] save tmp fail botID=%s tgID=%d err=%v", botID, tgID, err)
		out.Close()
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при сохранении голосового."))
		return
	}
	out.Close()
	defer os.Remove(path)

	log.Printf("[voice] saved to %s", path)

	// голос -> текст
	text, err := app.SpeechService.Transcribe(ctx, path)
	if err != nil {
		log.Printf("[voice] transcribe fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос."))
		return
	}
	log.Printf("[voice] transcribed: %q", text)

	if _, err := app.RecordService.AddText(ctx, botID, tgID, "user", text); err != nil {
		log.Printf("[voice] AddText user fail botID=%s tgID=%d err=%v", botID, tgID, err)
	}

	// GPT ответ
	reply, err := app.AiService.GetReply(ctx, botID, tgID, text, nil)
	if err != nil {
		log.Printf("[voice] ai fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при ответе."))
		return
	}
	log.Printf("[voice] gpt reply: %q", reply)

	// ответ -> голос
	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, reply, outVoice); err != nil {
		log.Printf("[voice] synth fail botID=%s tgID=%d err=%v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, reply))
		return
	}
	log.Printf("[voice] synthesized -> %s", outVoice)
	defer os.Remove(outVoice)

	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	if _, err := bot.Send(voice); err != nil {
		log.Printf("[voice] send fail: %v", err)
	} else {
		log.Printf("[voice] sent 🎤")
	}

	if _, err := app.RecordService.AddText(ctx, botID, tgID, "tutor", reply); err != nil {
		log.Printf("[voice] AddText tutor fail botID=%s tgID=%d err=%v", botID, tgID, err)
	}

	log.Printf("[voice] done botID=%s tgID=%d", botID, tgID)
}
