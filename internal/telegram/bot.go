package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// runBotLoop — главный цикл получения апдейтов
func (app *BotApp) runBotLoop(botID string, bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)
	log.Printf("[bot_loop] started botID=%s username=@%s", botID, bot.Self.UserName)

	for update := range updates {
		ctx := context.Background()
		tgID := extractTelegramID(update)
		if tgID == 0 {
			continue
		}

		log.Printf("[bot_loop] update received botID=%s tgID=%d", botID, tgID)

		status, err := app.SubscriptionService.GetStatus(ctx, botID, tgID)
		if err != nil {
			log.Printf("[bot_loop] getStatus fail botID=%s tgID=%d err=%v", botID, tgID, err)
			continue
		}

		switch {
		case update.Message != nil:
			app.handleMessage(ctx, botID, bot, update.Message.Chat.ID, tgID, status, update)

		case update.CallbackQuery != nil:
			app.handleCallback(ctx, botID, bot, update.CallbackQuery, status)
		}
	}
}

func extractTelegramID(u tgbotapi.Update) int64 {
	switch {
	case u.Message != nil && u.Message.From != nil:
		return u.Message.From.ID
	case u.CallbackQuery != nil && u.CallbackQuery.From != nil:
		return u.CallbackQuery.From.ID
	default:
		return 0
	}
}

func (app *BotApp) handleMessage(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	chatID, tgID int64, status string, update tgbotapi.Update) {

	switch status {

	case "none":
		log.Printf("[bot_loop] no subscription botID=%s tgID=%d → show menu", botID, tgID)
		menu := app.BuildSubscriptionMenu(ctx)
		text := app.BuildSubscriptionText()
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = menu
		bot.Send(msg)
		return

	case "pending":
		log.Printf("[bot_loop] pending botID=%s tgID=%d", botID, tgID)
		msg := tgbotapi.NewMessage(chatID, MsgPending)
		bot.Send(msg)
		return

	case "active":
		log.Printf("[bot_loop] active botID=%s tgID=%d", botID, tgID)

		msg := update.Message

		// --- голосовое ---
		if msg.Voice != nil {
			handleVoice(ctx, app, bot, botID, chatID, tgID, msg)
			return
		}

		// --- изображение ---
		if len(msg.Photo) > 0 {
			handlePhoto(ctx, app, bot, botID, chatID, tgID, msg)
			return
		}

		// --- текст ---
		if msg.Text == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "📎 Отправь текст, голос или фото."))
			return
		}

		reply, err := app.AiService.GetReply(ctx, botID, tgID, msg.Text)
		if err != nil {
			log.Printf("[bot_loop] ai reply fail: %v", err)
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке запроса."))
			return
		}

		outVoice := fmt.Sprintf("/tmp/reply_text_%d.mp3", tgID)
		if err := app.SpeechService.Synthesize(ctx, reply, outVoice); err == nil {
			voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
			bot.Send(voice)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, reply))
		}

	default:
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Неизвестный статус подписки."))
	}
}

// --- голос ---
func handleVoice(ctx context.Context, app *BotApp, bot *tgbotapi.BotAPI,
	botID string, chatID, tgID int64, msg *tgbotapi.Message) {

	fileID := msg.Voice.FileID
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("[bot_loop] get voice file fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить голосовое."))
		return
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[bot_loop] download voice fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при загрузке голосового."))
		return
	}
	defer resp.Body.Close()

	path := fmt.Sprintf("/tmp/%s.ogg", fileID)
	out, _ := os.Create(path)
	io.Copy(out, resp.Body)
	out.Close()

	text, err := app.SpeechService.Transcribe(ctx, path)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось распознать голос."))
		return
	}
	app.RecordService.AddText(ctx, botID, tgID, "user", text)

	reply, err := app.AiService.GetReply(ctx, botID, tgID, text)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при ответе."))
		return
	}

	outVoice := fmt.Sprintf("/tmp/reply_%s.mp3", fileID)
	if err := app.SpeechService.Synthesize(ctx, reply, outVoice); err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, reply))
		return
	}

	voice := tgbotapi.NewVoice(chatID, tgbotapi.FilePath(outVoice))
	bot.Send(voice)
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)
}

// --- фото ---
func handlePhoto(ctx context.Context, app *BotApp, bot *tgbotapi.BotAPI,
	botID string, chatID, tgID int64, msg *tgbotapi.Message) {

	photo := msg.Photo[len(msg.Photo)-1] // берем лучшее качество
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
	if err != nil {
		log.Printf("[bot_loop] get photo fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Не удалось получить фото."))
		return
	}

	url := file.Link(bot.Token)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[bot_loop] download photo fail: %v", err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка загрузки фото."))
		return
	}
	defer resp.Body.Close()

	tmpFile := fmt.Sprintf("/tmp/%s.jpg", photo.FileID)
	out, _ := os.Create(tmpFile)
	io.Copy(out, resp.Body)
	out.Close()

	f, _ := os.Open(tmpFile)
	defer f.Close()

	app.RecordService.AddImage(ctx, botID, tgID, "user", multipart.File(f), filepath.Base(tmpFile), "image/jpeg")

	urlMsg := fmt.Sprintf("📷 Пользователь прислал изображение: %s", url)
	reply, err := app.AiService.GetReply(ctx, botID, tgID, urlMsg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка обработки фото."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, reply))
	app.RecordService.AddText(ctx, botID, tgID, "tutor", reply)
}

func (app *BotApp) handleCallback(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	cb *tgbotapi.CallbackQuery, status string) {

	tgID := cb.From.ID
	chatID := cb.Message.Chat.ID
	log.Printf("[bot_loop] callback botID=%s tgID=%d data=%s", botID, tgID, cb.Data)

	switch status {
	case "none":
		paymentURL, err := app.SubscriptionService.Create(ctx, botID, tgID, cb.Data)
		if err != nil {
			log.Printf("[bot_loop] create payment fail botID=%s tgID=%d: %v", botID, tgID, err)
			bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка оформления"))
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при создании оплаты, попробуй позже.")
			bot.Send(msg)
			return
		}

		bot.Request(tgbotapi.NewCallback(cb.ID, "Заявка принята"))
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Заявка принята!\nДля завершения оплаты перейди по ссылке:\n%s", paymentURL))
		bot.Send(msg)

	case "pending", "active":
		bot.Request(tgbotapi.NewCallback(cb.ID, "Подписка уже оформлена"))
		msg := tgbotapi.NewMessage(chatID, MsgAlreadySubscribed)
		bot.Send(msg)
	}
}