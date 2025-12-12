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
		Model            *string
		TextStylePrompt  *string
		VoiceStylePrompt *string
		PhotoStylePrompt *string
		VoiceID          *string
		WelcomeText      *string
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
		PhotoStylePrompt: body.PhotoStylePrompt,
		WelcomeText:      body.WelcomeText,
		VoiceID:          body.VoiceID,
	}

	out, err := h.svc.Update(r.Context(), in)
	if err != nil {
		http.Error(w, "failed to update bot config", 500)
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}

// POST /bots/{bot_id}/welcome-video
func (h *Handler) UploadWelcomeVideo(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", 400)
		return
	}
	defer file.Close()

	url, err := h.svc.UploadWelcomeVideo(r.Context(), botID, file, header.Filename)
	if err != nil {
		http.Error(w, "upload failed", 500)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"url": url,
	})
}
