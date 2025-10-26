package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type TariffHandler struct {
	service ports.TariffService
}

func NewTariffHandler(s ports.TariffService) *TariffHandler {
	return &TariffHandler{service: s}
}

func (h *TariffHandler) List(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get tariffs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(plans)
}
