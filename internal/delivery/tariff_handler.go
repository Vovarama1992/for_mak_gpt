package delivery

import (
	"encoding/json"
	"net/http"

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

	created, err := h.svc.Create(r.Context(), &input)
	if err != nil {
		http.Error(w, "failed to create tariff", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(created)
}
