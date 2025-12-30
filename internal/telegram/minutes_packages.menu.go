package telegram

import (
	"context"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BuildMinutePackagesMenu — показывает кнопки с пакетами минут
func (app *BotApp) BuildMinutePackagesMenu(
	ctx context.Context,
	botID string,
) tgbotapi.InlineKeyboardMarkup {

	pkgs, err := app.MinutePackageService.ListAll(ctx)
	if err != nil || len(pkgs) == 0 {
		log.Printf("[minute_packages] load fail: %v", err)
		btn := tgbotapi.NewInlineKeyboardButtonData("Ошибка загрузки пакетов", "none")
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, p := range pkgs {
		if !p.Active {
			continue
		}

		label := p.Name + " — " + strconv.Itoa(p.Minutes) + " мин / " + formatRUB(p.Price)
		data := "pkg_" + strconv.FormatInt(p.ID, 10)

		btn := tgbotapi.NewInlineKeyboardButtonData(label, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	if len(rows) == 0 {
		btn := tgbotapi.NewInlineKeyboardButtonData("Нет доступных пакетов", "none")
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
