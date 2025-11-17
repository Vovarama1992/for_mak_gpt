package delivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Role       string `json:"role"`
		Text       string `json:"text"`
		BotID      string `json:"bot_id"`
	}

	// первая попытка — обычный JSON
	if err := json.Unmarshal(body, &req); err != nil {
		// fallback: патчим переносы строк, если Make вставил их "вживую"
		clean := bytes.ReplaceAll(body, []byte("\r"), []byte(""))
		clean = bytes.ReplaceAll(clean, []byte("\n"), []byte("\\n"))
		if err2 := json.Unmarshal(clean, &req); err2 != nil {
			http.Error(w, "invalid json after fallback: "+err2.Error(), http.StatusBadRequest)
			return
		}
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

	origin := r.Header.Get("Origin")
	isAdmin := strings.Contains(origin, "aifull.ru") // определяем, что запрос пришёл с админки

	for i := range history {
		if !isAdmin {
			continue // GPT — не трогаем
		}
		if history[i].ImageURL == nil || *history[i].ImageURL == "" {
			continue
		}

		raw := *history[i].ImageURL
		if strings.Contains(raw, "%2F") {
			parts := strings.SplitN(raw, botID+"%2F", 2)
			if len(parts) == 2 {
				fixed := fmt.Sprintf("https://aifull.ru/api/images/%s/%s",
					botID,
					strings.ReplaceAll(parts[1], "%2F", "/"),
				)
				history[i].ImageURL = &fixed
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
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
