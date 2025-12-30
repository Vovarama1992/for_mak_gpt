package telegram

import (
	"context"
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// helper — достаём число в начале строки
func extractNumber(s string) (int, bool) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (app *BotApp) ShowClassPicker(
	ctx context.Context,
	botID string,
	bot *tgbotapi.BotAPI,
	tgID int64,
	chatID int64,
) {
	// список классов
	list, err := app.ClassService.ListClasses(ctx, botID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список классов"))
		return
	}

	// -----------------------------------------------------
	// СОРТИРОВКА: сперва цифры по возрастанию, затем строки
	// -----------------------------------------------------
	sort.Slice(list, func(i, j int) bool {
		ai, okA := extractNumber(list[i].Grade)
		bj, okB := extractNumber(list[j].Grade)

		if okA && okB {
			// оба начинаются с числа
			return ai < bj
		}
		if okA && !okB {
			// число выше строки
			return true
		}
		if !okA && okB {
			// строка идёт после чисел
			return false
		}

		// оба строковые → по алфавиту
		return list[i].Grade < list[j].Grade
	})
	// -----------------------------------------------------

	text := "📚 Выбор класса\n\n"

	// inline-кнопки
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, c := range list {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s класс", c.Grade),
				fmt.Sprintf("set_class_%d", c.ID),
			),
		})
	}

	menu := tgbotapi.NewInlineKeyboardMarkup(rows...)

	out := tgbotapi.NewMessage(chatID, text)
	out.ReplyMarkup = menu
	bot.Send(out)
}
