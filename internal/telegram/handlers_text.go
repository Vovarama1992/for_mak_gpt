package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) handleText(ctx context.Context, botID string, bot *tgbotapi.BotAPI,
	msg *tgbotapi.Message, tgID int64) {

	chatID := msg.Chat.ID
	text := msg.Text

	// GPT-ответ (история старее → новый текст → супер-системный промпт)
	reply, err := app.AiService.GetReply(ctx, botID, tgID, text, nil)
	if err != nil {
		log.Printf("[text] ai reply fail botID=%s tgID=%d: %v", botID, tgID, err)
		bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Ошибка при обработке запроса."))
		return
	}

	// Отправляем ответ
	bot.Send(tgbotapi.NewMessage(chatID, reply))

	// Пишем в историю после ответа — сперва user, затем tutor
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "user", text); err != nil {
		log.Printf("[text] AddText user fail botID=%s tgID=%d err=%v", botID, tgID, err)
	}
	if _, err := app.RecordService.AddText(ctx, botID, tgID, "tutor", reply); err != nil {
		log.Printf("[text] AddText tutor fail botID=%s tgID=%d err=%v", botID, tgID, err)
	}
}
