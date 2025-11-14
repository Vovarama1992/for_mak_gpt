package delivery

import (
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes регистрирует все HTTP-маршруты приложения
func RegisterRoutes(
	r chi.Router,
	h *RecordHandler,
	hSubs *SubscriptionHandler,
	hTariff *TariffHandler,
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

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Post("/record/image", h.AddImageRecord)

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Get("/history/{telegram_id}", h.GetHistory)

	r.With(httputil.RecoverMiddleware).Get("/admin/users", h.ListUsers)

	// --- подписки ---
	r.With(httputil.RecoverMiddleware).Post("/subscribe/create", hSubs.Create)
	r.With(httputil.RecoverMiddleware).Post("/subscribe/activate", hSubs.Activate)
	r.With(httputil.RecoverMiddleware).Get("/subscribe/status/{telegram_id}", hSubs.GetStatus)
	r.With(httputil.RecoverMiddleware).Get("/admin/subscriptions", hSubs.ListAll)

	// --- тарифные планы ---
	r.With(httputil.RecoverMiddleware).Get("/tariffs", hTariff.List)
	r.With(httputil.RecoverMiddleware).Post("/tariffs", hTariff.Create)
}
