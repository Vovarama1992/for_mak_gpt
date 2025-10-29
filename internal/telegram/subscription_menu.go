package telegram

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BuildSubscriptionMenu — формирует клавиатуру с тарифами.
func (app *BotApp) BuildSubscriptionMenu(ctx context.Context) tgbotapi.InlineKeyboardMarkup {
	tariffs, err := app.TariffService.ListAll(ctx)
	if err != nil {
		log.Printf("[subscription_menu] list fail: %v", err)
		return errorMenu("Ошибка загрузки тарифов")
	}
	if len(tariffs) == 0 {
		return errorMenu("Нет доступных тарифов")
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tariffs {
		label := fmt.Sprintf("%s — %s", t.Name, formatRUB(t.Price))
		btn := tgbotapi.NewInlineKeyboardButtonData(label, t.Code)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// errorMenu — заглушка, если тарифы не удалось получить.
func errorMenu(text string) tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonData(text, "none")
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
}

// formatRUB форматирует цену: 199 → "199 ₽", 199.5 → "199.50 ₽"
func formatRUB(p float64) string {
	if p == math.Trunc(p) {
		return fmt.Sprintf("%.0f ₽", p)
	}
	// до 2 знаков, без хвостовых нулей после обрезки
	s := fmt.Sprintf("%.2f", p)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s + " ₽"
}
