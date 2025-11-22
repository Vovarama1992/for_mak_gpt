package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ShowVoiceMinutesScreen — экран остатка минут (команда /voice)
func (app *BotApp) ShowVoiceMinutesScreen(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	tgID int64,
	chatID int64,
) {
	sub, err := app.SubscriptionService.Get(ctx, botID, tgID)
	if err != nil || sub == nil {
		out := tgbotapi.NewMessage(chatID, "Не удалось получить данные подписки. Попробуй позже.")
		bot.Send(out)
		return
	}

	// просто минуты целиком, без секунд
	minutes := int(sub.VoiceMinutes)

	text := fmt.Sprintf(
		"🎤 Голосовые минуты\n\nОсталось: %d мин.\n\nКогда минуты закончатся, ты сможешь продолжать обучение текстом.",
		minutes,
	)

	btn := tgbotapi.NewInlineKeyboardButtonData("Пополнить голос", "buy_voice")
	menu := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{btn},
	)

	out := tgbotapi.NewMessage(chatID, text)
	out.ReplyMarkup = menu
	bot.Send(out)
}
