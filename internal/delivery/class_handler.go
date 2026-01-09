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

func getBotID(r *http.Request) string {
	return r.Header.Get("X-Bot-ID")
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

func (h *ClassHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BotID string `json:"bot_id"`
		Grade string `json:"grade"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if body.BotID == "" {
		http.Error(w, "bot_id required", 400)
		return
	}

	out, err := h.svc.CreateClass(
		r.Context(),
		body.BotID,
		body.Grade,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}

func (h *ClassHandler) UpdateClass(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "class_id")
	id, _ := strconv.Atoi(idStr)

	var body struct {
		BotID string `json:"bot_id"`
		Grade string `json:"grade"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if body.BotID == "" {
		http.Error(w, "bot_id required", 400)
		return
	}

	if err := h.svc.UpdateClass(
		r.Context(),
		body.BotID,
		id,
		body.Grade,
	); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClassHandler) DeleteClass(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "class_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid class_id", 400)
		return
	}

	if err := h.svc.DeleteClassByID(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//
// ----------------------
//   ПРОМПТЫ
// ----------------------
//

func (h *ClassHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	botID := getBotID(r)

	cidStr := chi.URLParam(r, "class_id")
	cid, _ := strconv.Atoi(cidStr)

	p, err := h.svc.GetPromptByClassID(r.Context(), botID, cid)
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
func (h *ClassHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	botID := getBotID(r)

	cidStr := chi.URLParam(r, "class_id")
	classID, _ := strconv.Atoi(cidStr)

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	out, err := h.svc.CreatePrompt(r.Context(), botID, classID, body.Prompt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(out)
}

// PATCH /prompts/{prompt_id}
func (h *ClassHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	botID := getBotID(r)

	pidStr := chi.URLParam(r, "prompt_id")
	pid, _ := strconv.Atoi(pidStr)

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.svc.UpdatePrompt(r.Context(), botID, pid, body.Prompt); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

// DELETE /prompts/{prompt_id}
func (h *ClassHandler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	botID := getBotID(r)

	pidStr := chi.URLParam(r, "prompt_id")
	pid, _ := strconv.Atoi(pidStr)

	if err := h.svc.DeletePrompt(r.Context(), botID, pid); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
