package delivery

import (
	"encoding/json"
	"net/http"

	tr "github.com/Vovarama1992/make_ziper/internal/textrules"
)

type TextRuleHandler struct {
	repo tr.Repo
}

func NewTextRuleHandler(repo tr.Repo) *TextRuleHandler {
	return &TextRuleHandler{repo: repo}
}

//
// ----------------------
//   LETTER RULES
// ----------------------
//

// GET /text-rules/letters
func (h *TextRuleHandler) ListLetterRules(w http.ResponseWriter, r *http.Request) {
	out, err := h.repo.ListLetterRules(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(out)
}

// POST /text-rules/letters
// body: { "from": "ё", "to": "е" }
func (h *TextRuleHandler) AddLetterRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.repo.AddLetterRule(r.Context(), body.From, body.To); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

// DELETE /text-rules/letters
// body: { "from": "ё" }
func (h *TextRuleHandler) DeleteLetterRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.repo.DeleteLetterRule(r.Context(), body.From); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

//
// ----------------------
//   WORD RULES
// ----------------------
//

// GET /text-rules/words
func (h *TextRuleHandler) ListWordRules(w http.ResponseWriter, r *http.Request) {
	out, err := h.repo.ListWordRules(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(out)
}

// POST /text-rules/words
// body: { "from": "AI", "to": "ИИ" }
func (h *TextRuleHandler) AddWordRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.repo.AddWordRule(r.Context(), body.From, body.To); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}

// DELETE /text-rules/words
// body: { "from": "AI" }
func (h *TextRuleHandler) DeleteWordRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := h.repo.DeleteWordRule(r.Context(), body.From); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(204)
}
