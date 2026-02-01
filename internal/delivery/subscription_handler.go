package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	service ports.SubscriptionService
}

func NewSubscriptionHandler(service ports.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// POST /subscribe/create
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BotID      string `json:"bot_id"`
		TelegramID int64  `json:"telegram_id"`
		PlanCode   string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.BotID == "" || req.TelegramID == 0 || req.PlanCode == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	paymentURL, err := h.service.Create(r.Context(), req.BotID, req.TelegramID, req.PlanCode)
	if err != nil {
		http.Error(w, "failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "created",
		"payment_url": paymentURL,
	})
}

// POST /subscribe/activate
func (h *SubscriptionHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	data, ok := body["Data"].(map[string]any)
	if !ok {
		http.Error(w, "missing Data", 400)
		return
	}

	paymentType, _ := data["payment_type"].(string)
	botID, _ := data["bot_id"].(string)

	telegramID := int64(data["telegram_id"].(float64))
	packageID := int64(data["package_id"].(float64))

	switch paymentType {

	case "minute_package":
		if err := h.service.AddMinutesFromPackage(
			r.Context(),
			botID,
			telegramID,
			packageID,
		); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	case "subscription":
		invoiceID, _ := body["InvoiceId"].(string)
		if err := h.service.Activate(r.Context(), invoiceID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	default:
		http.Error(w, "unknown payment_type", 400)
		return
	}

	w.Write([]byte("OK"))
}

// GET /subscribe/status/{telegram_id}?bot_id=...
func (h *SubscriptionHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	botID := r.URL.Query().Get("bot_id")
	tidStr := chi.URLParam(r, "telegram_id")

	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetStatus(r.Context(), botID, tid)
	if err != nil {
		http.Error(w, "failed to get status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": status})
}

func (h *SubscriptionHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list subscriptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type dto struct {
		ID           int64    `json:"id"`
		BotID        string   `json:"bot_id"`
		TelegramID   int64    `json:"telegram_id"`
		PlanName     string   `json:"plan_name"`
		Status       string   `json:"status"`
		StartedAt    *string  `json:"started_at"`
		ExpiresAt    *string  `json:"expires_at"`
		VoiceMinutes *float64 `json:"voice_minutes"`
	}
	out := make([]dto, 0, len(subs))

	for _, s := range subs {
		var started, expires *string

		if s.StartedAt != nil {
			str := s.StartedAt.Format(time.RFC3339)
			started = &str
		}
		if s.ExpiresAt != nil {
			str := s.ExpiresAt.Format(time.RFC3339)
			expires = &str
		}

		var vm *float64
		if s.VoiceMinutes > 0 {
			vm = &s.VoiceMinutes
		}

		out = append(out, dto{
			ID:           s.ID,
			BotID:        s.BotID,
			TelegramID:   s.TelegramID,
			PlanName:     s.PlanName,
			Status:       s.Status,
			StartedAt:    started,
			ExpiresAt:    expires,
			VoiceMinutes: vm,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	tidStr := chi.URLParam(r, "telegram_id")

	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), botID, telegramID); err != nil {
		http.Error(w, "failed to delete subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /subscribe/{id}
func (h *SubscriptionHandler) UpdateLimits(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	var raw struct {
		Status       string  `json:"status"`
		ExpiresAt    string  `json:"expires_at"`
		VoiceMinutes float64 `json:"voice_minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	t, err := time.Parse("2006-01-02", raw.ExpiresAt)
	if err != nil {
		http.Error(w, "invalid expires_at format", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateLimits(
		r.Context(),
		subID,
		raw.Status,
		&t,
		raw.VoiceMinutes,
	); err != nil {
		http.Error(w, "failed to update subscription: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}
