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
	mainKB tgbotapi.ReplyKeyboardMarkup,
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
	go app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedMinutes)

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	resp, err := http.Get(file.Link(bot.Token))
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
	io.Copy(out, resp.Body)
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

	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, _ := bot.Send(thinking)

	reply, err := app.AiService.GetReply(ctx, botID, tgID, "voice", text, nil)
	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки запроса.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		m := tgbotapi.NewMessage(chatID, reply)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer os.Remove(outVoice)

	if durSec, err := speech.AudioDuration(outVoice); err == nil {
		go app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, durSec/60.0)
	}

	// === ФИКС КЛАВИАТУРЫ ===

	// 1) удалить thinking
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	// 2) отправить voice
	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	bot.Send(voice)

	// 3) сохранить историю
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// 4) финальное сообщение с клавиатурой (последнее)
	keep := tgbotapi.NewMessage(chatID, "\u200b")
	keep.ReplyMarkup = mainKB
	bot.Send(keep)
}
