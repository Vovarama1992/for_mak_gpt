package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) ShowClassPicker(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	tgID int64,
	chatID int64,
) {
	// текущий

	// список классов
	list, err := app.ClassService.ListClasses(ctx)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список классов"))
		return
	}

	text := "📚 Выбор класса\n\n"

	// inline-кнопки
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, c := range list {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d класс", c.Grade),
				fmt.Sprintf("set_class_%d", c.ID),
			),
		})
	}

	menu := tgbotapi.NewInlineKeyboardMarkup(rows...)

	out := tgbotapi.NewMessage(chatID, text)
	out.ReplyMarkup = menu
	bot.Send(out)
}
