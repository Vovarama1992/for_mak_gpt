package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	cl "github.com/Vovarama1992/make_ziper/internal/classes"
	"github.com/go-chi/chi/v5"
)

type ClassHandler struct {
	svc cl.ClassService
}

func NewClassHandler(svc cl.ClassService) *ClassHandler {
	return &ClassHandler{svc: svc}
}

// GET /class-prompts
func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListPrompts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(out)
}

// POST /class-prompts
func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p cl.ClassPrompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.svc.CreatePrompt(context.Background(), &p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// PATCH /class-prompts/{class}
func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	classStr := chi.URLParam(r, "class")
	class, _ := strconv.Atoi(classStr)

	var p cl.ClassPrompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// находим ID по class
	existing, err := h.svc.GetPromptByClass(r.Context(), class)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if existing == nil {
		http.Error(w, "not found", 404)
		return
	}

	p.ID = existing.ID
	p.Class = class

	if err := h.svc.UpdatePrompt(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// DELETE /class-prompts/{class}
func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	classStr := chi.URLParam(r, "class")
	class, _ := strconv.Atoi(classStr)

	existing, err := h.svc.GetPromptByClass(r.Context(), class)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if existing == nil {
		http.Error(w, "not found", 404)
		return
	}

	if err := h.svc.DeletePrompt(r.Context(), existing.ID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
