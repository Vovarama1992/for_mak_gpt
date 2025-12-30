package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (app *BotApp) BuildMainKeyboard(status string) tgbotapi.ReplyKeyboardMarkup {
	first := "🟢 Начать урок"
	if status == "active" {
		first = "🟢 Продолжить урок"
	}

	row1 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(first),
	)

	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("💳 Тарифы"),
		tgbotapi.NewKeyboardButton("📦 Пакеты минут"),
	)

	row3 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❓ Помощь"),
	)

	row4 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🗑 Очистить историю"),
		tgbotapi.NewKeyboardButton("🧹 Сбросить настройки"),
	)

	kb := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
	kb.ResizeKeyboard = true
	return kb
}
