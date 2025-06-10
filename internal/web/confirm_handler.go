package web

import (
	"net/http"
	"strings"
)

func (h *Handler) confirmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !h.validateCSRF(w, r) {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/confirm/")
	code := strings.TrimSpace(r.FormValue("code"))
	ip := extractIP(r)

	err := h.store.Confirm(id, code, ip)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}

	w.Write([]byte("Secret unlocked. Recipient can now view the secret."))
}
