package prompts

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list prompts", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Prompt == "" {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updated, err := h.svc.Update(r.Context(), botID, body.Prompt)
	if err != nil {
		http.Error(w, "failed to update prompt", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(updated)
}
