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
		http.Error(w, "failed to list bot configs", 500)
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// GET /bots/{bot_id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", 400)
		return
	}

	cfg, err := h.svc.Get(r.Context(), botID)
	if err != nil {
		http.Error(w, "bot not found", 404)
		return
	}

	_ = json.NewEncoder(w).Encode(cfg)
}

// PATCH /bots/{bot_id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", 400)
		return
	}

	var body struct {
		Model            *string `json:"model"`
		TextStylePrompt  *string `json:"text_style_prompt"`
		VoiceStylePrompt *string `json:"voice_style_prompt"`
		VoiceID          *string `json:"voice_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	in := &UpdateInput{
		BotID:            botID,
		Model:            body.Model,
		TextStylePrompt:  body.TextStylePrompt,
		VoiceStylePrompt: body.VoiceStylePrompt,
		VoiceID:          body.VoiceID,
	}

	out, err := h.svc.Update(r.Context(), in)
	if err != nil {
		http.Error(w, "failed to update bot config", 500)
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}
