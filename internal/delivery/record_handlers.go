package delivery

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Vovarama1992/go-utils/logger"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/go-chi/chi/v5"
)

type RecordHandler struct {
	recordService ports.RecordService
	log           *logger.ZapLogger
}

func NewRecordHandler(recordService ports.RecordService, log *logger.ZapLogger) *RecordHandler {
	return &RecordHandler{
		recordService: recordService,
		log:           log,
	}
}

func (h *RecordHandler) AddTextRecordJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// заменяем сырые переводы строк, если они мешают JSON
	clean := bytes.ReplaceAll(body, []byte("\r"), []byte(""))
	clean = bytes.ReplaceAll(clean, []byte("\n"), []byte("\\n"))

	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Role       string `json:"role"`
		Text       string `json:"text"`
		BotID      string `json:"bot_id"`
	}

	if err := json.Unmarshal(clean, &req); err != nil {
		http.Error(w, "invalid json after patch: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.BotID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddText(r.Context(), req.BotID, req.TelegramID, req.Role, req.Text)
	if err != nil {
		http.Error(w, "failed to save text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *RecordHandler) AddTextRecordForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}

	telegramID, _ := strconv.ParseInt(r.FormValue("telegram_id"), 10, 64)
	role := r.FormValue("role")
	text := r.FormValue("text")
	botID := r.FormValue("bot_id")

	if telegramID == 0 || text == "" || botID == "" {
		http.Error(w, "missing telegram_id, bot_id or text", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddText(r.Context(), botID, telegramID, role, text)
	if err != nil {
		http.Error(w, "failed to save text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *RecordHandler) AddImageRecord(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "invalid multipart: "+err.Error(), http.StatusBadRequest)
		return
	}

	telegramStr := r.FormValue("telegram_id")
	role := r.FormValue("role")
	botID := r.FormValue("bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tid, err := strconv.ParseInt(telegramStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	id, err := h.recordService.AddImage(r.Context(), botID, tid, role, file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "failed to upload image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *RecordHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	tidStr := chi.URLParam(r, "telegram_id")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	botID := r.URL.Query().Get("bot_id")
	if botID == "" {
		http.Error(w, "missing bot_id", http.StatusBadRequest)
		return
	}

	history, err := h.recordService.GetHistory(r.Context(), botID, tid)
	if err != nil {
		h.log.Log(logger.LogEntry{Level: "error", Message: "db error", Error: err})
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *RecordHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.recordService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "failed to list users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
