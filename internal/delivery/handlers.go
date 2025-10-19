package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Vovarama1992/go-utils/logger"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	recordService ports.RecordService
	log           *logger.ZapLogger
}

func NewHandler(recordService ports.RecordService, log *logger.ZapLogger) *Handler {
	return &Handler{
		recordService: recordService,
		log:           log,
	}
}

func (h *Handler) AddTextRecord(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TelegramID int64  `json:"telegram_id"`
		Role       string `json:"role"`
		Text       string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.log.Log(logger.LogEntry{Level: "warn", Message: "invalid json", Error: err})
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddText(r.Context(), payload.TelegramID, payload.Role, payload.Text)
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "error", Message: "failed to create record", Error: err})
		http.Error(w, "failed to create record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) AddImageRecord(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "warn", Message: "invalid multipart", Error: err})
		http.Error(w, "invalid multipart: "+err.Error(), http.StatusBadRequest)
		return
	}

	telegramStr := r.FormValue("telegram_id")
	role := r.FormValue("role")
	file, header, err := r.FormFile("file")
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "warn", Message: "missing file", Error: err})
		http.Error(w, "missing file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tid, err := strconv.ParseInt(telegramStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddImage(r.Context(), tid, role, file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "error", Message: "failed to upload image", Error: err})
		http.Error(w, "failed to upload image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	tidStr := chi.URLParam(r, "telegram_id")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	history, err := h.recordService.GetHistory(r.Context(), tid)
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "error", Message: "db error", Error: err})
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
