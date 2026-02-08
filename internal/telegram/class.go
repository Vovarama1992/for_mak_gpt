package telegram

import (
	"context"
	"fmt"
	"sort"

	"github.com/Vovarama1992/make_ziper/internal/classes"

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
	cfg, err := app.BotsService.Get(ctx, botID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка конфигурации бота"))
		return
	}

	label := cfg.ClassLabel

	list, err := app.ClassService.ListClasses(ctx, botID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось получить список"))
		return
	}

	var filtered []*classes.Class
	for _, c := range list {
		if c.BotID == botID {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("%s не найдены", label)))
		return
	}

	sort.Slice(filtered, func(i, j int) bool {
		ai, okA := extractNumber(filtered[i].Grade)
		bj, okB := extractNumber(filtered[j].Grade)

		if okA && okB {
			return ai < bj
		}
		if okA {
			return true
		}
		if okB {
			return false
		}
		return filtered[i].Grade < filtered[j].Grade
	})

	rows := [][]tgbotapi.InlineKeyboardButton{}

	for _, c := range filtered {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				c.Grade,
				fmt.Sprintf("set_class_%d", c.ID),
			),
		))
	}

	msg := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("📚 Выбор: %s", label),
	)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	bot.Send(msg)
}
