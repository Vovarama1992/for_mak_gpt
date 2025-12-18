package delivery

import (
	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/ports"
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
	authSvc ports.AuthService,
) {
	// --- auth ---
	r.With(httputil.RecoverMiddleware).
		Post("/auth/login", hAuth.Login)

	// --- protected ---
	r.Route("/", func(pr chi.Router) {
		pr.Use(
			httputil.RecoverMiddleware,
			AuthMiddleware(authSvc),
		)

		// --- записи ---
		pr.Post("/record/text/user", h.AddTextRecordJSON)
		pr.Post("/record/text/tutor", h.AddTextRecordForm)
		pr.Delete("/records", h.DeleteAll)
		pr.Get("/history/{telegram_id}", h.GetHistory)
		pr.Get("/users", h.ListUsers)

		// --- подписки ---
		pr.Post("/subscribe/create", hSubs.Create)
		pr.Post("/subscribe/activate", hSubs.Activate)
		pr.Get("/subscribe/status/{telegram_id}", hSubs.GetStatus)
		pr.Get("/subscriptions", hSubs.ListAll)
		pr.Delete("/subscribe/{bot_id}/{telegram_id}", hSubs.Delete)

		// --- тарифы ---
		pr.Get("/tariffs", hTariff.List)
		pr.Post("/tariffs", hTariff.Create)
		pr.Put("/tariffs/{id}", hTariff.Update)
		pr.Delete("/tariffs/{id}", hTariff.Delete)

		// --- боты ---
		pr.Get("/bots", hBots.List)
		pr.Get("/bots/{bot_id}", hBots.Get)
		pr.Patch("/bots/{bot_id}", hBots.Update)
		pr.Post("/bots/{bot_id}/welcome-video", hBots.UploadWelcomeVideo)

		// --- пакеты минут ---
		pr.Get("/minute-packages", hPkg.List)
		pr.Post("/minute-packages", hPkg.Create)
		pr.Get("/minute-packages/{id}", hPkg.Get)
		pr.Patch("/minute-packages/{id}", hPkg.Update)
		pr.Delete("/minute-packages/{id}", hPkg.Delete)

		// --- классы ---
		pr.Get("/classes", hClass.ListClasses)
		pr.Post("/classes", hClass.CreateClass)
		pr.Patch("/classes/{class_id}", hClass.UpdateClass)
		pr.Delete("/classes/{class_id}", hClass.DeleteClass)
		pr.Get("/classes/{class_id}/prompts", hClass.GetPrompt)
		pr.Post("/classes/{class_id}/prompts", hClass.CreatePrompt)
		pr.Patch("/prompts/{prompt_id}", hClass.UpdatePrompt)
		pr.Delete("/prompts/{prompt_id}", hClass.DeletePrompt)
	})
}
