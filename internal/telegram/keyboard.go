package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (app *BotApp) BuildMainKeyboard(botID, status string) tgbotapi.ReplyKeyboardMarkup {
	first := "🟢 Начать"

	if botID == "assistant" {
		first = "🟢 Начать"
	}

	if status == "active" {
		if botID == "assistant" {
			first = "🟢 Продолжить диалог"
		} else {
			first = "🟢 Продолжить"
		}
	}

	row1 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(first),
	)

	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("💳 Тарифы"),
		tgbotapi.NewKeyboardButton("📦 Остаток минут"),
	)

	row3 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❓ Помощь"),
	)

	row4 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🗑 Очистить диалог"),
		tgbotapi.NewKeyboardButton("🧹 Начать заново"),
	)

	kb := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
	kb.ResizeKeyboard = true
	return kb
}
