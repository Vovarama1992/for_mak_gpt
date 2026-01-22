package delivery

import (
	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	h *RecordHandler,
	hSubs *SubscriptionHandler,
	hTariff *TariffHandler,
	hBots *bots.Handler,
	hPkg *MinutePackageHandler,
	hClass *ClassHandler,
	hAuth *AuthHandler,
	hTextRules *TextRuleHandler,
) {
	// --- auth ---
	r.With(httputil.RecoverMiddleware).
		Post("/auth/login", hAuth.Login)

	// --- записи ---
	r.With(httputil.RecoverMiddleware).
		Post("/record/text/user", h.AddTextRecordJSON)

	r.With(httputil.RecoverMiddleware).
		Post("/record/text/tutor", h.AddTextRecordForm)

	r.With(httputil.RecoverMiddleware).
		Delete("/records", h.DeleteAll)

	r.With(httputil.RecoverMiddleware).
		Get("/history/{telegram_id}", h.GetHistory)

	r.With(httputil.RecoverMiddleware).
		Get("/users", h.ListUsers)

	// --- подписки ---
	r.With(httputil.RecoverMiddleware).
		Post("/subscribe/create", hSubs.Create)

	r.With(httputil.RecoverMiddleware).
		Post("/subscribe/activate", hSubs.Activate)

	r.With(httputil.RecoverMiddleware).
		Get("/subscribe/status/{telegram_id}", hSubs.GetStatus)

	r.With(httputil.RecoverMiddleware).
		Get("/subscriptions", hSubs.ListAll)

	r.With(httputil.RecoverMiddleware).
		Delete("/subscribe/{bot_id}/{telegram_id}", hSubs.Delete)

	// --- тарифы ---
	r.With(httputil.RecoverMiddleware).
		Get("/tariffs", hTariff.List)

	r.With(httputil.RecoverMiddleware).
		Post("/tariffs", hTariff.Create)

	r.With(httputil.RecoverMiddleware).
		Put("/tariffs/{id}", hTariff.Update)

	r.With(httputil.RecoverMiddleware).
		Delete("/tariffs/{id}", hTariff.Delete)

	// --- боты ---
	r.With(httputil.RecoverMiddleware).
		Get("/bots", hBots.List)

	r.With(httputil.RecoverMiddleware).
		Get("/bots/{bot_id}", hBots.Get)

	r.With(httputil.RecoverMiddleware).
		Patch("/bots/{bot_id}", hBots.Update)
	r.With(httputil.RecoverMiddleware).
		Delete("/bots/{bot_id}", hBots.Delete)

	r.With(httputil.RecoverMiddleware).
		Post("/bots", hBots.Create)

	r.With(httputil.RecoverMiddleware).
		Post("/bots/{bot_id}/welcome-video", hBots.UploadWelcomeVideo)

	// --- пакеты минут ---
	r.With(httputil.RecoverMiddleware).
		Get("/minute-packages", hPkg.List)

	r.With(httputil.RecoverMiddleware).
		Post("/minute-packages", hPkg.Create)

	r.With(httputil.RecoverMiddleware).
		Get("/minute-packages/{id}", hPkg.Get)

	r.With(httputil.RecoverMiddleware).
		Patch("/minute-packages/{id}", hPkg.Update)

	r.With(httputil.RecoverMiddleware).
		Delete("/minute-packages/{id}", hPkg.Delete)

	// --- классы ---
	r.With(httputil.RecoverMiddleware).
		Get("/classes", hClass.ListClasses)

	r.With(httputil.RecoverMiddleware).
		Post("/classes", hClass.CreateClass)

	r.With(httputil.RecoverMiddleware).
		Patch("/classes/{class_id}", hClass.UpdateClass)

	r.With(httputil.RecoverMiddleware).
		Delete("/classes/{class_id}", hClass.DeleteClass)

	r.With(httputil.RecoverMiddleware).
		Get("/classes/{class_id}/prompts", hClass.GetPrompt)

	r.With(httputil.RecoverMiddleware).
		Post("/classes/{class_id}/prompts", hClass.CreatePrompt)

	r.With(httputil.RecoverMiddleware).
		Patch("/prompts/{prompt_id}", hClass.UpdatePrompt)

	r.With(httputil.RecoverMiddleware).
		Delete("/prompts/{prompt_id}", hClass.DeletePrompt)

	// --- text rules ---
	r.With(httputil.RecoverMiddleware).
		Get("/text-rules/letters", hTextRules.ListLetterRules)

	r.With(httputil.RecoverMiddleware).
		Post("/text-rules/letters", hTextRules.AddLetterRule)

	r.With(httputil.RecoverMiddleware).
		Delete("/text-rules/letters", hTextRules.DeleteLetterRule)

	r.With(httputil.RecoverMiddleware).
		Get("/text-rules/words", hTextRules.ListWordRules)

	r.With(httputil.RecoverMiddleware).
		Post("/text-rules/words", hTextRules.AddWordRule)

	r.Patch("/subscribe/{id}", hSubs.UpdateLimits)

	r.With(httputil.RecoverMiddleware).
		Delete("/text-rules/words", hTextRules.DeleteWordRule)
}
