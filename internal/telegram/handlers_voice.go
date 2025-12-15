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
		log.Printf("[TG SEND] type=message reason=voice_not_allowed kb=1")
		m := tgbotapi.NewMessage(chatID, "🔇 В этом тарифе голос временно недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	usedMinutes := float64(msg.Voice.Duration) / 60.0
	go func() {
		ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedMinutes)
		log.Printf("[voice] charge stt_minutes=%.4f ok=%v err=%v", usedMinutes, ok, err)
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
		log.Printf("[TG SEND] type=message reason=get_file_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[TG SEND] type=message reason=download_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка при загрузке голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, err := os.Create(path)
	if err != nil {
		log.Printf("[TG SEND] type=message reason=create_tmp_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		log.Printf("[TG SEND] type=message reason=save_tmp_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка сохранения голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	out.Close()
	defer os.Remove(path)

	text, err := app.SpeechService.Transcribe(ctx, botID, path)
	if err != nil {
		log.Printf("[TG SEND] type=message reason=stt_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	app.RecordService.AddText(ctx, botID, tgID, "user", text)
	log.Printf("[voice] stt_text=%q", text)

	// индикатор — НЕ УДАЛЯЕМ в эксперименте
	log.Printf("[TG SEND] type=message reason=thinking kb=1")
	thinking := tgbotapi.NewMessage(chatID, "🤖 AI думает…")
	thinking.ReplyMarkup = mainKB
	sentThinking, sendErr := bot.Send(thinking)
	log.Printf("[TG SENT] thinking msgID=%d err=%v", sentThinking.MessageID, sendErr)

	reply, err := app.AiService.GetReply(ctx, botID, tgID, "voice", text, nil)
	if err != nil {
		log.Printf("[TG SEND] type=message reason=gpt_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка AI.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {
		log.Printf("[TG SEND] type=message reason=tts_fail kb=1 err=%v", err)
		m := tgbotapi.NewMessage(chatID, reply)
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer os.Remove(outVoice)

	// списание TTS — НЕ РЕЖЕМ
	if durSec, err := speech.AudioDuration(outVoice); err == nil {
		usedReplyMinutes := durSec / 60.0
		log.Printf("[voice] tts_duration_sec=%.3f tts_minutes=%.4f", durSec, usedReplyMinutes)
		go func() {
			ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, usedReplyMinutes)
			log.Printf("[voice] charge tts_minutes=%.4f ok=%v err=%v", usedReplyMinutes, ok, err)
			if err != nil {
				app.ErrorNotify.Notify(ctx, botID, err,
					fmt.Sprintf("Ошибка списания TTS минут: tg=%d", tgID))
				return
			}
			if !ok {
				log.Printf("[voice] async: no voice minutes left for TTS tgID=%d", tgID)
			}
		}()
	} else {
		log.Printf("[voice] tts_duration_err=%v", err)
	}

	// отправляем voice
	log.Printf("[TG SEND] type=voice kb=0")
	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	_, vErr := bot.Send(voice)
	log.Printf("[TG SENT] voice err=%v", vErr)

	// сохраняем историю
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)

	// финальное сообщение с клавой — обязательно
	log.Printf("[TG SEND] type=message reason=keep kb=1")
	keep := tgbotapi.NewMessage(chatID, "\u200b")
	keep.ReplyMarkup = mainKB
	_, kErr := bot.Send(keep)
	log.Printf("[TG SENT] keep err=%v", kErr)

	log.Printf("[voice] end botID=%s tgID=%d (thinking_kept msgID=%d)", botID, tgID, sentThinking.MessageID)
}
