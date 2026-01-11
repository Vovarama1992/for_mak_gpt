package telegram

import (
	"context"
	"log"
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

	// -------- ОСТАТОК МИНУТ (ИНФО) --------
	sub, err := app.SubscriptionService.Get(ctx, botID, tgID)
	if err == nil && sub != nil {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🎧 Остаток минут: "+strconv.FormatFloat(sub.VoiceMinutes, 'f', 2, 64),
					"noop",
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Пакеты минут:", "noop"),
			),
		)
	}

	// -------- ПАКЕТЫ МИНУТ (КЛИКАБЕЛЬНЫЕ) --------
	pkgs, err := app.MinutePackageService.ListAll(ctx)
	if err != nil {
		log.Printf("[minute_packages] load fail: %v", err)
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ошибка загрузки пакетов", "noop"),
			),
		)
	}

	for _, p := range pkgs {
		if p.BotID != botID || !p.Active {
			continue
		}

		label := p.Name + " — " +
			strconv.Itoa(p.Minutes) + " мин / " + formatRUB(p.Price)

		btn := tgbotapi.NewInlineKeyboardButtonData(
			label,
			"pkg_"+strconv.FormatInt(p.ID, 10),
		)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	if len(rows) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Нет доступных пакетов", "noop"),
			),
		)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
