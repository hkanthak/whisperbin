package web

import (
	"fmt"
	"net/http"
	"strings"
)

func (h *Handler) confirmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !h.validateCSRF(w, r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/confirm/")
	code := strings.TrimSpace(r.FormValue("code"))

	ip := extractIP(r)
	err := h.store.Confirm(id, code, ip)
	if err != nil {
		http.Error(w, "Invalid code or already unlocked", http.StatusForbidden)
		return
	}

	fmt.Fprintln(w, "Secret unlocked. Recipient can now view the secret.")
}
