package delivery

import (
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

//
// ----------------------
//   КЛАССЫ
// ----------------------
//

// GET /classes
func (h *ClassHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.svc.ListClasses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(classes)
}

// POST /classes
// body: { "grade": 1 }
func (h *ClassHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Grade int `json:"grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	out, err := h.svc.CreateClass(r.Context(), body.Grade)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(out)
}

//
// ----------------------
//   ПРОМПТЫ ДЛЯ КЛАССА
// ----------------------
//

func (h *ClassHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	cidStr := chi.URLParam(r, "class_id")
	cid, _ := strconv.Atoi(cidStr)

	p, err := h.svc.GetPromptByClassID(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if p == nil {
		http.Error(w, "not found", 404)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// POST /classes/{class_id}/prompts
// body: { "prompt": "..." }
func (h *ClassHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	cidStr := chi.URLParam(r, "class_id")
	classID, _ := strconv.Atoi(cidStr)

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	out, err := h.svc.CreatePrompt(r.Context(), classID, body.Prompt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(out)
}

// PATCH /prompts/{prompt_id}
// body: { "prompt": "..." }
func (h *ClassHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	pidStr := chi.URLParam(r, "prompt_id")
	pid, _ := strconv.Atoi(pidStr)

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.svc.UpdatePrompt(r.Context(), pid, body.Prompt); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

// DELETE /prompts/{prompt_id}
func (h *ClassHandler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	pidStr := chi.URLParam(r, "prompt_id")
	pid, _ := strconv.Atoi(pidStr)

	if err := h.svc.DeletePrompt(r.Context(), pid); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
