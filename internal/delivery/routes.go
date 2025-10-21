package delivery

import (
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {

	// пользователь — JSON
	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(100, time.Minute),
	).Post("/record/text/user", h.AddTextRecordJSON)

	// тьютор (GPT) — form-urlencoded
	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(100, time.Minute),
	).Post("/record/text/tutor", h.AddTextRecordForm)

	// изображение
	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Post("/record/image", h.AddImageRecord)

	// история
	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Get("/history/{telegram_id}", h.GetHistory)
}
