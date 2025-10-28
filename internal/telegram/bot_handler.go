package telegram

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// BotHandler — HTTP-прослойка для взаимодействия с Telegram-ботами
type BotHandler struct {
	dispatcher *Dispatcher
}

func NewBotHandler(dispatcher *Dispatcher) *BotHandler {
	return &BotHandler{dispatcher: dispatcher}
}

// POST /telegram/show-menu
func (h *BotHandler) ShowMenu(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BotID      string `json:"bot_id"`
		TelegramID string `json:"telegram_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.BotID == "" || req.TelegramID == "" {
		http.Error(w, "missing bot_id or telegram_id", http.StatusBadRequest)
		return
	}

	tid, err := strconv.ParseInt(req.TelegramID, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	go h.dispatcher.CheckAndShowMenu(r.Context(), req.BotID, tid)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
