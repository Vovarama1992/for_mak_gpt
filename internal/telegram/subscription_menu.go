package telegram

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *BotApp) BuildSubscriptionMenu(
	ctx context.Context,
	botID string,
) tgbotapi.InlineKeyboardMarkup {

	tariffs, err := app.TariffService.ListAll(ctx)
	if err != nil {
		log.Printf("[subscription_menu] list fail err=%v", err)
		return errorMenu("Ошибка загрузки тарифов")
	}

	log.Printf(
		"[subscription_menu] fetched tariffs total=%d botID=%s",
		len(tariffs),
		botID,
	)

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, t := range tariffs {
		log.Printf(
			"[subscription_menu] check tariff id=%d botID=%s code=%s isTrial=%v",
			t.ID,
			t.BotID,
			t.Code,
			t.IsTrial,
		)

		if t.BotID != botID {
			log.Printf(
				"[subscription_menu] skip tariff code=%s reason=botID_mismatch tariffBotID=%s",
				t.Code,
				t.BotID,
			)
			continue
		}

		if t.IsTrial {
			log.Printf(
				"[subscription_menu] skip tariff code=%s reason=is_trial",
				t.Code,
			)
			continue
		}

		voice := "∞ мин голоса"
		if t.VoiceMinutes < 9_000_000 {
			voice = fmt.Sprintf("%d мин голоса", int(t.VoiceMinutes))
		}

		label := fmt.Sprintf(
			"%s — %s (%s, %s)",
			t.Name,
			formatRUB(t.Price),
			minutesToDays(t.DurationMinutes),
			voice,
		)

		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					label,
					"sub:"+t.Code,
				),
			),
		)

		log.Printf(
			"[subscription_menu] added tariff code=%s label=%q",
			t.Code,
			label,
		)
	}

	log.Printf(
		"[subscription_menu] result rows=%d botID=%s",
		len(rows),
		botID,
	)

	if len(rows) == 0 {
		log.Printf(
			"[subscription_menu] EMPTY result botID=%s",
			botID,
		)
		return errorMenu("Нет доступных тарифов")
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (app *BotApp) BuildSubscriptionText(
	ctx context.Context,
	botID string,
) string {

	// 🔹 тянем конфиг бота (ДАЖЕ ЕСЛИ СЕРВИСА ЕЩЁ НЕТ)
	cfg, err := app.BotsService.Get(ctx, botID)
	if err == nil && cfg != nil && cfg.TariffText != nil && *cfg.WelcomeText != "" {
		return *cfg.TariffText
	}

	// fallback — если текста нет
	return "🎓 Тарифы AI-репетитора\n\n" +
		"Выберите подходящий тариф ниже ⬇️"
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
