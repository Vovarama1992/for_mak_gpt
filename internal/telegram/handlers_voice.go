package telegram

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

	if !app.checkVoiceAllowed(ctx, botID, tgID) {
		m := tgbotapi.NewMessage(chatID, "🔇 В этом тарифе голос недоступен.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	// ===== ПРОВЕРКА ЛИМИТА (ЕДИНСТВЕННОЕ ИЗМЕНЕНИЕ) =====

	used := float64(msg.Voice.Duration) / 60.0
	ok, err := app.SubscriptionService.UseVoiceMinutes(ctx, botID, tgID, used)
	if err != nil || !ok {
		cfg, _ := app.BotsService.Get(ctx, botID)

		text := "Недостаточно голосовых минут."
		if cfg != nil && cfg.NoVoiceMinutesText != nil {
			if t := strings.TrimSpace(*cfg.NoVoiceMinutesText); t != "" {
				text = t
			}
		}

		m := tgbotapi.NewMessage(chatID, text)
		m.ReplyMarkup = app.BuildMinutePackagesMenu(ctx, botID)
		bot.Send(m)
		return
	}

	// ===== ДАЛЬШЕ ФАЙЛ БЕЗ ИЗМЕНЕНИЙ =====

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	resp, err := http.Get(file.Link(bot.Token))
	if err != nil {
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки голосового.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, _ := os.Create(path)
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
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка AI.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}

	processed, err := app.TextRuleService.Process(ctx, reply)
	if err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки текста.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	reply = processed

	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, botID, reply, outVoice); err != nil {
		bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))
		m := tgbotapi.NewMessage(chatID, "⚠️ Ошибка озвучки.")
		m.ReplyMarkup = mainKB
		bot.Send(m)
		return
	}
	defer os.Remove(outVoice)

	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentThinking.MessageID))

	anchor := tgbotapi.NewMessage(chatID, "🎧 Ответ голосом:")
	anchor.ReplyMarkup = mainKB
	bot.Send(anchor)

	bot.Send(tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice)))

	textMsg := tgbotapi.NewMessage(chatID, reply)
	textMsg.ReplyMarkup = mainKB
	bot.Send(textMsg)

	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)
}
