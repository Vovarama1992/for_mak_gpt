package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	if req.BotID == "" || req.TelegramID == 0 {
		http.Error(w, "missing bot_id or telegram_id", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), req.BotID, req.TelegramID, req.PlanCode); err != nil {
		http.Error(w, "failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"status": "created"})
}

// POST /subscribe/activate
func (h *SubscriptionHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// пробуем вытащить bot_id и telegram_id из metadata
	var botID string
	var telegramID int64

	if meta, ok := body["metadata"].(map[string]any); ok {
		if v, ok := meta["bot_id"].(string); ok {
			botID = v
		}
		if v, ok := meta["telegram_id"].(string); ok {
			if id, err := strconv.ParseInt(v, 10, 64); err == nil {
				telegramID = id
			}
		}
	}

	// если нет metadata — fallback на старый формат
	if botID == "" || telegramID == 0 {
		botID, _ = body["bot_id"].(string)
		if v, ok := body["telegram_id"].(float64); ok {
			telegramID = int64(v)
		}
	}

	if botID == "" || telegramID == 0 {
		http.Error(w, "missing bot_id or telegram_id", http.StatusBadRequest)
		return
	}

	if err := h.service.Activate(r.Context(), botID, telegramID); err != nil {
		http.Error(w, "failed to activate: "+err.Error(), http.StatusInternalServerError)
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

// GET /admin/subscriptions
func (h *SubscriptionHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list subscriptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(subs)
}
