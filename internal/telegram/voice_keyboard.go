package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func buildVoiceKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🕒 Остаток минут"),
		),
	)
}
