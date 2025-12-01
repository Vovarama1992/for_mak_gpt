package delivery

import (
	"time"

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
) {
	// --- записи ---
	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(100, time.Minute),
	).Post("/record/text/user", h.AddTextRecordJSON)

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(100, time.Minute),
	).Post("/record/text/tutor", h.AddTextRecordForm)

	r.With(httputil.RecoverMiddleware).Delete("/records", h.DeleteAll)

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Get("/history/{telegram_id}", h.GetHistory)

	r.With(httputil.RecoverMiddleware).Get("/users", h.ListUsers)

	// --- подписки ---
	r.With(httputil.RecoverMiddleware).Post("/subscribe/create", hSubs.Create)
	r.With(httputil.RecoverMiddleware).Post("/subscribe/activate", hSubs.Activate)
	r.With(httputil.RecoverMiddleware).Get("/subscribe/status/{telegram_id}", hSubs.GetStatus)
	r.With(httputil.RecoverMiddleware).Get("/subscriptions", hSubs.ListAll)

	// --- тарифные планы ---
	r.With(httputil.RecoverMiddleware).Get("/tariffs", hTariff.List)
	r.With(httputil.RecoverMiddleware).Post("/tariffs", hTariff.Create)
	r.With(httputil.RecoverMiddleware).Put("/tariffs/{id}", hTariff.Update)
	r.With(httputil.RecoverMiddleware).Delete("/tariffs/{id}", hTariff.Delete)

	// --- бот-конфиги ---
	r.With(httputil.RecoverMiddleware).Get("/bots", hBots.List)
	r.With(httputil.RecoverMiddleware).Get("/bots/{bot_id}", hBots.Get)
	r.With(httputil.RecoverMiddleware).Patch("/bots/{bot_id}", hBots.Update)

	// --- пакеты минут ---
	r.With(httputil.RecoverMiddleware).Get("/minute-packages", hPkg.List)
	r.With(httputil.RecoverMiddleware).Post("/minute-packages", hPkg.Create)
	r.With(httputil.RecoverMiddleware).Get("/minute-packages/{id}", hPkg.Get)
	r.With(httputil.RecoverMiddleware).Patch("/minute-packages/{id}", hPkg.Update)
	r.With(httputil.RecoverMiddleware).Delete("/minute-packages/{id}", hPkg.Delete)

	// --- class prompts ---
	r.Get("/classes", hClass.ListClasses)
	r.Post("/classes", hClass.CreateClass)

	r.Get("/classes/{class_id}/prompts", hClass.GetPrompt)
	r.Post("/classes/{class_id}/prompts", hClass.CreatePrompt)

	r.Patch("/prompts/{prompt_id}", hClass.UpdatePrompt)
	r.Delete("/prompts/{prompt_id}", hClass.DeletePrompt)
}
