package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// helpKeyboard — клавиатура режима помощи
func helpKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⬅️ Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}
