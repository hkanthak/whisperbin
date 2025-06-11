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
		h.renderError(w, http.StatusForbidden, "Invalid Request", "	Invalid request. Please try again.")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/confirm/")
	code := strings.TrimSpace(r.FormValue("code"))
	ip := extractIP(r)

	err := h.store.Confirm(id, code, ip)
	if err != nil {
		h.renderError(w, http.StatusForbidden, "Invalid Request", "Invalid confirmation code or expired link.")
		return
	}

	h.renderSuccess(w, "Secret unlocked", "Recipient can now view the secret.")
}
