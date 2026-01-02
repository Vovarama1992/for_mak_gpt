package telegram

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) BuildSubscriptionMenu(
	ctx context.Context,
	botID string,
) tgbotapi.InlineKeyboardMarkup {

	tariffs, err := app.TariffService.ListAll(ctx)
	if err != nil {
		log.Printf("[subscription_menu] list fail: %v", err)
		return errorMenu("Ошибка загрузки тарифов")
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, t := range tariffs {
		if t.BotID != botID || t.IsTrial {
			continue
		}

		label := fmt.Sprintf("%s — %s", t.Name, formatRUB(t.Price))

		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					label,
					fmt.Sprintf("sub_preview:%d", t.ID),
				),
			),
		)
	}

	if len(rows) == 0 {
		return errorMenu("Нет доступных тарифов")
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (app *BotApp) BuildSubscriptionText() string {
	return "🎓 Тарифы AI-репетитора\n\n" +
		"Выберите тариф, чтобы посмотреть описание ⬇️"
}

func (app *BotApp) HandleTariffPreview(
	ctx context.Context,
	botID string,
	cb *tgbotapi.CallbackQuery,
) {
	idStr := strings.TrimPrefix(cb.Data, "sub_preview:")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return
	}

	t, err := app.TariffService.GetByID(ctx, botID, id)
	if err != nil {
		return
	}

	voice := "∞ мин голоса"
	if t.VoiceMinutes < 9_000_000 {
		voice = fmt.Sprintf("%d мин голоса", int(t.VoiceMinutes))
	}

	text := fmt.Sprintf(
		"%s — %s\n\n"+
			"🕒 %s\n"+
			"🎤 %s\n\n"+
			"%s\n\n"+
			"Подключить тариф?",
		t.Name,
		formatRUB(t.Price),
		minutesToDays(t.DurationMinutes),
		voice,
		t.Description,
	)

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Подключить",
				fmt.Sprintf("sub_confirm:%s", t.Code),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "sub_back"),
		),
	)

	bot := app.bots[botID]

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		text,
		kb,
	)

	bot.Send(msg)
}

func errorMenu(text string) tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonData(text, "noop")
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
}

func minutesToDays(minutes int) string {
	if minutes <= 0 {
		return "0 дн"
	}
	days := minutes / (24 * 60)
	if days <= 0 {
		return "< 1 дн"
	}
	return fmt.Sprintf("%d дн", days)
}

func formatRUB(p float64) string {
	if p == math.Trunc(p) {
		return fmt.Sprintf("%.0f ₽", p)
	}
	s := fmt.Sprintf("%.2f", p)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s + " ₽"
}
