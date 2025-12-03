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

func (app *BotApp) handleVoice(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message,
	tgID int64,
	mainKB tgbotapi.ReplyKeyboardMarkup, // ← добавили
) {
	chatID := msg.Chat.ID
	fileID := msg.Voice.FileID

	log.Printf("[voice] start botID=%s tgID=%d fileID=%s", botID, tgID, fileID)

	if !app.checkVoiceAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "🔇 В этом тарифе голос временно недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
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
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка при загрузке голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, err := os.Create(path)
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка сохранения голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	out.Close()
	defer os.Remove(path)

	text, err := app.SpeechService.Transcribe(ctx, botID, path)
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	app.RecordService.AddText(ctx, botID, tgID, "user", text)

	// === индикатор ===
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	// === GPT ===
	reply, err := app.AiService.GetReply(ctx, botID, tgID, "voice", text, nil)
	if err != nil {
		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)

		m := tgbotapi.NewMessage(chatID, reply)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	// === синтез ===
	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {
		del := tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID)
		bot.Request(del)

		m := tgbotapi.NewMessage(chatID, reply)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer os.Remove(outVoice)

	// === списание TTS ===
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
	// ReplyMarkup в Voice нельзя вставить, поэтому СРАЗУ после него отправляем пустое сообщение с клавой
	bot.Send(voice)

	// сохраняем историю
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// удаляем индикатор
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	// ← именно ЭТО фиксирует меню после voice → GPT → audio
	keep := tgbotapi.NewMessage(chatID, " ")
	keep.ReplyMarkup = mainKB
	bot.Send(keep)
}
