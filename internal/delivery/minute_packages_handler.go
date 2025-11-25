package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	mp "github.com/Vovarama1992/make_ziper/internal/minutes_packages"
)

type MinutePackageHandler struct {
	svc mp.MinutePackageService
}

func NewMinutePackageHandler(svc mp.MinutePackageService) *MinutePackageHandler {
	return &MinutePackageHandler{svc: svc}
}

// GET /minute-packages
func (h *MinutePackageHandler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(out)
}

// GET /minute-packages/{id}
func (h *MinutePackageHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	out, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if out == nil {
		http.Error(w, "not found", 404)
		return
	}

	json.NewEncoder(w).Encode(out)
}

// POST /minute-packages
func (h *MinutePackageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var pkg mp.MinutePackage
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.svc.Create(context.Background(), &pkg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(pkg)
}

// PATCH /minute-packages/{id}
func (h *MinutePackageHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var pkg mp.MinutePackage
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	pkg.ID = id

	if err := h.svc.Update(r.Context(), &pkg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(pkg)
}

// DELETE /minute-packages/{id}
func (h *MinutePackageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
