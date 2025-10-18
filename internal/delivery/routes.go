package delivery

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	withRecover := func(nextHandler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
		return httputil.RecoverHandler(nextHandler)
	}

	withRate := func(rps int, per time.Duration) func(nextHandler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
		return httputil.NewRateLimiter(rps, per)
	}

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
