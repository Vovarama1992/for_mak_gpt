package delivery

import (
	"encoding/json"
	"io"
	"log"
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
		"keyboard": []map[string]any{
			{
				"type": "button",
				"text": "📄 Документы",
				"data": "docs",
			},
			{
				"type": "url",
				"text": "Оплатить",
				"url":  paymentURL,
			},
		},
	})
}

// POST /subscribe/activate
func (h *SubscriptionHandler) Activate(w http.ResponseWriter, r *http.Request) {
	log.Println("[PAY][YK] webhook hit")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("[PAY][YK] read body error:", err)
		http.Error(w, "bad body", 400)
		return
	}

	log.Println("[PAY][YK] raw body:", string(body))

	var notif struct {
		Type   string `json:"type"`
		Event  string `json:"event"`
		Object struct {
			ID       string `json:"id"`
			Status   string `json:"status"`
			Paid     bool   `json:"paid"`
			Metadata struct {
				BotID      string `json:"bot_id"`
				TelegramID string `json:"telegram_id"`
				PackageID  string `json:"package_id"`
				Type       string `json:"payment_type"`
				PlanCode   string `json:"plan_code"`
				InvoiceID  string `json:"invoice_id"`
			} `json:"metadata"`
		} `json:"object"`
	}

	if err := json.Unmarshal(body, &notif); err != nil {
		log.Println("[PAY][YK] json decode error:", err)
		http.Error(w, "bad json", 400)
		return
	}

	log.Printf("[PAY][YK] event=%s status=%s paid=%v paymentID=%s\n",
		notif.Event,
		notif.Object.Status,
		notif.Object.Paid,
		notif.Object.ID,
	)

	if notif.Event != "payment.succeeded" {
		log.Println("[PAY][YK] ignore event:", notif.Event)
		w.WriteHeader(200)
		return
	}

	meta := notif.Object.Metadata
	log.Printf("[PAY][YK] meta: %+v\n", meta)

	tgID, _ := strconv.ParseInt(meta.TelegramID, 10, 64)
	pkgID, _ := strconv.ParseInt(meta.PackageID, 10, 64)

	switch meta.Type {

	case "minute_package":
		log.Println("[PAY][YK] add minutes start")

		if err := h.service.AddMinutesFromPackage(
			r.Context(),
			meta.BotID,
			tgID,
			pkgID,
		); err != nil {
			log.Println("[PAY][YK] add minutes error:", err)
			http.Error(w, err.Error(), 500)
			return
		}

		log.Println("[PAY][YK] minutes added OK")

	case "subscription":
		log.Println("[PAY][YK] activate subscription start")

		invoiceID := notif.Object.Metadata.InvoiceID
		log.Println("[PAY][YK] invoice_id:", invoiceID)

		if invoiceID == "" {
			log.Println("[PAY][YK] missing invoice_id in metadata")
			http.Error(w, "missing invoice_id", 400)
			return
		}

		if err := h.service.Activate(
			r.Context(),
			invoiceID,
		); err != nil {
			log.Println("[PAY][YK] activate sub error:", err)
			http.Error(w, err.Error(), 500)
			return
		}

		log.Println("[PAY][YK] subscription activated")

	default:
		log.Println("[PAY][YK] unknown payment_type:", meta.Type)
	}

	w.WriteHeader(200)
	w.Write([]byte("ok"))
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
