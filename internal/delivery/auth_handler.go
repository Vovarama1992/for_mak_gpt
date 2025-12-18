package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type AuthHandler struct {
	auth ports.AuthService
}

func NewAuthHandler(auth ports.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, err := h.auth.Login(r.Context(), req.Password)
	if err != nil {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
