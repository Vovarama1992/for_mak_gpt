package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type TariffHandler struct {
	svc ports.TariffService
}

func NewTariffHandler(svc ports.TariffService) *TariffHandler {
	return &TariffHandler{svc: svc}
}

func (h *TariffHandler) List(w http.ResponseWriter, r *http.Request) {

	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list tariffs", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(items)
}

func (h *TariffHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input ports.TariffPlan
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if input.BotID == "" {
		http.Error(w, "bot_id required", http.StatusBadRequest)
		return
	}

	created, err := h.svc.Create(r.Context(), &input)
	if err != nil {
		http.Error(w, "failed to create tariff", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(created)
}

func (h *TariffHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input ports.TariffPlan
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if input.BotID == "" {
		http.Error(w, "bot_id required", http.StatusBadRequest)
		return
	}

	input.ID = id

	updated, err := h.svc.Update(r.Context(), &input)
	if err != nil {
		http.Error(w, "failed to update tariff", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(updated)
}

func (h *TariffHandler) Delete(w http.ResponseWriter, r *http.Request) {
	botID := r.URL.Query().Get("bot_id")
	if botID == "" {
		http.Error(w, "bot_id required", http.StatusBadRequest)
		return
	}

	id, err := extractIDFromURL(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), botID, id); err != nil {
		http.Error(w, "failed to delete tariff", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractIDFromURL(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(parts[len(parts)-1])
}
