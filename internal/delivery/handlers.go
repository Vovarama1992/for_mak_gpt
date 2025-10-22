package delivery

import (
	"encoding/json"
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

func (h *Handler) AddTextRecordJSON(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Role       string `json:"role"`
		Text       string `json:"text"`
		BotID      string `json:"bot_id"`
	}

	log.Printf("[AddTextRecordJSON] request from %s", r.RemoteAddr)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AddTextRecordJSON] invalid json: %v", err)
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	var botIDPtr *string
	if req.BotID != "" {
		botIDPtr = &req.BotID
	}

	log.Printf("[AddTextRecordJSON] parsed: telegram_id=%d role=%q bot_id=%q text_len=%d",
		req.TelegramID, req.Role, req.BotID, len(req.Text))

	id, err := h.recordService.AddText(r.Context(), botIDPtr, req.TelegramID, req.Role, req.Text)
	if err != nil {
		log.Printf("[AddTextRecordJSON] failed to save text: %v", err)
		http.Error(w, "failed to save text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[AddTextRecordJSON] success: id=%v", id)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *Handler) AddTextRecordForm(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AddTextRecordForm] request from %s", r.RemoteAddr)

	if err := r.ParseForm(); err != nil {
		log.Printf("[AddTextRecordForm] invalid form: %v", err)
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}

	telegramID, _ := strconv.ParseInt(r.FormValue("telegram_id"), 10, 64)
	role := r.FormValue("role")
	text := r.FormValue("text")
	botID := r.FormValue("bot_id")

	var botIDPtr *string
	if botID != "" {
		botIDPtr = &botID
	}

	log.Printf("[AddTextRecordForm] parsed form: telegram_id=%d role=%q bot_id=%q text_len=%d",
		telegramID, role, botID, len(text))

	if telegramID == 0 || text == "" {
		log.Printf("[AddTextRecordForm] invalid input: telegram_id=%d text_len=%d", telegramID, len(text))
		http.Error(w, "missing telegram_id or text", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddText(r.Context(), botIDPtr, telegramID, role, text)
	if err != nil {
		log.Printf("[AddTextRecordForm] failed to save text: %v", err)
		http.Error(w, "failed to save text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[AddTextRecordForm] success: id=%v", id)
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
	botID := r.FormValue("bot_id")

	var botIDPtr *string
	if botID != "" {
		botIDPtr = &botID
	}

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

	id, err := h.recordService.AddImage(r.Context(), botIDPtr, tid, role, file, header.Filename, header.Header.Get("Content-Type"))
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

	botID := r.URL.Query().Get("bot_id")
	var botIDPtr *string
	if botID != "" {
		botIDPtr = &botID
	}

	history, err := h.recordService.GetHistory(r.Context(), botIDPtr, tid)
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "error", Message: "db error", Error: err})
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
