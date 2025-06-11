package web

import (
	"fmt"
	"net/http"
	"strings"
)

func (h *Handler) recipientHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")

	secret, err := h.store.Get(id)
	if err != nil {
		h.renderError(w, http.StatusNotFound, "Not Found", "Secret not found or expired.")
		return
	}

	if secret.Unlocked {
		text, err := h.store.DecryptSecretText(secret)
		if err != nil {
			h.renderError(w, http.StatusInternalServerError, "Internal Error", "An unexpected error occurred.")
			return
		}
		h.store.Delete(id)
		h.templates.ExecuteTemplate(w, "show.html", text)
		return
	}

	data := struct {
		ID   string
		Code string
	}{
		ID:   id,
		Code: secret.Code,
	}
	h.templates.ExecuteTemplate(w, "waiting.html", data)
}

func (h *Handler) statusHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/status/")

	waiting, err := h.store.IsWaiting(id)
	if err != nil {
		http.Error(w, "not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"waiting": %t}`, waiting)
}
