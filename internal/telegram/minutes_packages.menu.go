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

	// -------- ОСТАТОК МИНУТ (ИНФО, НЕ ПАКЕТ) --------
	sub, err := app.SubscriptionService.Get(ctx, botID, tgID)
	if err == nil && sub != nil {
		label := "🎧 Остаток минут: " +
			strconv.FormatFloat(sub.VoiceMinutes, 'f', -1, 64)

		// noop — чтобы было некликабельно логически
		btn := tgbotapi.NewInlineKeyboardButtonData(label, "noop")
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

		// визуальный разделитель
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("────────────", "noop"),
		))
	}

	// -------- ПАКЕТЫ МИНУТ --------
	pkgs, err := app.MinutePackageService.ListAll(ctx)
	if err != nil {
		log.Printf("[minute_packages] load fail: %v", err)
		btn := tgbotapi.NewInlineKeyboardButtonData("Ошибка загрузки пакетов", "noop")
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
	}

	for _, p := range pkgs {
		if p.BotID != botID || !p.Active {
			continue
		}

		label := p.Name + " — " +
			strconv.Itoa(p.Minutes) + " мин / " + formatRUB(p.Price)

		data := "pkg_" + strconv.FormatInt(p.ID, 10)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, data)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	if len(rows) == 0 {
		btn := tgbotapi.NewInlineKeyboardButtonData("Нет доступных пакетов", "noop")
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
