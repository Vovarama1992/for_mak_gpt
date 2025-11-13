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
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Юкасса всегда передаёт payment_id в поле object.id
	object, ok := body["object"].(map[string]any)
	if !ok {
		http.Error(w, "missing object", http.StatusBadRequest)
		return
	}
	paymentID, _ := object["id"].(string)
	if paymentID == "" {
		http.Error(w, "missing payment_id", http.StatusBadRequest)
		return
	}

	if err := h.service.Activate(r.Context(), paymentID); err != nil {
		http.Error(w, "failed to activate subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": "activated"})
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
		ID         int64   `json:"id"`
		BotID      string  `json:"bot_id"`
		TelegramID int64   `json:"telegram_id"`
		PlanName   string  `json:"plan_name"`
		Status     string  `json:"status"`
		StartedAt  *string `json:"started_at"`
		ExpiresAt  *string `json:"expires_at"`
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

		out = append(out, dto{
			ID:         s.ID,
			BotID:      s.BotID,
			TelegramID: s.TelegramID,
			PlanName:   s.PlanName,
			Status:     s.Status,
			StartedAt:  started,
			ExpiresAt:  expires,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
