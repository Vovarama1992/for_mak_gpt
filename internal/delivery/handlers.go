package delivery

import (
	"encoding/json"
	"io"
	"log"
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[AddTextRecord] failed to read body: %v", err)
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("[AddTextRecord] raw body: %q", string(body))
	log.Printf("[AddTextRecord] content-type: %s", r.Header.Get("Content-Type"))

	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Role       string `json:"role"`
		Text       string `json:"text"`
	}

	// пробуем JSON
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[AddTextRecord] not JSON: %v", err)
		if err := r.ParseForm(); err == nil {
			req.TelegramID, _ = strconv.ParseInt(r.FormValue("telegram_id"), 10, 64)
			req.Role = r.FormValue("role")
			req.Text = r.FormValue("text")
			log.Printf("[AddTextRecord] parsed as form: telegram_id=%d, role=%q, text=%q",
				req.TelegramID, req.Role, req.Text)
		} else {
			log.Printf("[AddTextRecord] form parse error: %v", err)
		}
	} else {
		log.Printf("[AddTextRecord] parsed as JSON: telegram_id=%d, role=%q, text=%q",
			req.TelegramID, req.Role, req.Text)
	}

	if req.TelegramID == 0 || req.Text == "" {
		log.Printf("[AddTextRecord] invalid input: telegram_id=%d text=%q", req.TelegramID, req.Text)
		http.Error(w, "invalid input: missing telegram_id or text", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddText(r.Context(), req.TelegramID, req.Role, req.Text)
	if err != nil {
		log.Printf("[AddTextRecord] failed to save text: %v", err)
		http.Error(w, "failed to save text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[AddTextRecord] success: id=%v", id)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
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
