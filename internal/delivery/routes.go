package delivery

import (
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(100, time.Minute),
	).Post("/record/text", h.AddTextRecord)

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Post("/record/image", h.AddImageRecord)

	r.With(
		httputil.RecoverMiddleware,
		httputil.NewRateLimiter(60, time.Minute),
	).Get("/history/{telegram_id}", h.GetHistory)
}
