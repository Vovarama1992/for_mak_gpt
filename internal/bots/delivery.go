package bots

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GET /bots
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list bot configs", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// GET /bots/{bot_id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	cfg, err := h.svc.Get(r.Context(), botID)
	if err != nil {
		http.Error(w, "bot not found", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(cfg)
}

// PATCH /bots/{bot_id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	var body struct {
		Model       *string `json:"model"`
		StylePrompt *string `json:"style_prompt"`
		VoiceID     *string `json:"voice_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	in := &UpdateInput{
		BotID:       botID,
		Model:       body.Model,
		StylePrompt: body.StylePrompt,
		VoiceID:     body.VoiceID,
	}

	updated, err := h.svc.Update(r.Context(), in)
	if err != nil {
		http.Error(w, "failed to update bot config", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(updated)
}
