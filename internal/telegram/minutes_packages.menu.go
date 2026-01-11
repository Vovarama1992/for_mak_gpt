package telegram

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BuildMinutePackagesMenu — показывает кнопки с пакетами минут + остаток
// BuildMinutePackagesMenu — пакеты минут + ИНФО-СТРОКА С ОСТАТКОМ
func (app *BotApp) BuildMinutePackagesMenu(
	ctx context.Context,
	botID string,
	tgID int64,
) tgbotapi.InlineKeyboardMarkup {

	var rows [][]tgbotapi.InlineKeyboardButton

	pkgs, err := app.MinutePackageService.ListAll(ctx)
	if err != nil {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	for _, p := range pkgs {
		if p.BotID != botID || !p.Active {
			continue
		}

		label := fmt.Sprintf(
			"%s — %d мин / %s",
			p.Name,
			p.Minutes,
			formatRUB(p.Price),
		)

		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					label,
					"pkg_"+strconv.FormatInt(p.ID, 10),
				),
			),
		)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
