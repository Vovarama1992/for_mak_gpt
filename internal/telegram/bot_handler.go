package telegram

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type BotHandler struct {
	app *BotApp
}

func NewBotHandler(app *BotApp) *BotHandler {
	return &BotHandler{app: app}
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

	log.Printf("[ShowMenu] request received: bot_id=%s, telegram_id=%s", req.BotID, req.TelegramID)

	if req.BotID == "" || req.TelegramID == "" {
		http.Error(w, "missing bot_id or telegram_id", http.StatusBadRequest)
		return
	}

	tid, err := strconv.ParseInt(req.TelegramID, 10, 64)
	if err != nil {
		http.Error(w, "invalid telegram_id", http.StatusBadRequest)
		return
	}

	go func() {
		log.Printf("[ShowMenu] calling CheckSubscriptionAndShowMenu for bot_id=%s, telegram_id=%d", req.BotID, tid)
		h.app.CheckSubscriptionAndShowMenu(r.Context(), req.BotID, tid)
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
