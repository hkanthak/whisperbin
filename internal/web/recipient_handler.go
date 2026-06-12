package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) recipientHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")

	secret, err := h.store.Get(id)
	if err != nil {
		h.renderError(w, http.StatusNotFound, "Not Found", "Secret not found or expired.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if !secret.Unlocked {
			data := struct {
				ID   string
				Code string
			}{
				ID:   id,
				Code: secret.Code,
			}
			h.templates.ExecuteTemplate(w, "waiting.html", data)
			return
		}
		h.renderReveal(w, id)
	case http.MethodPost:
		if !secret.Unlocked {
			h.renderError(w, http.StatusNotFound, "Not Found", "Secret not found or expired.")
			return
		}
		if !h.validateCSRF(w, r) {
			h.renderError(w, http.StatusForbidden, "Forbidden", "Invalid CSRF token. Please reload the page and try again.")
			return
		}
		text, err := h.store.DecryptSecretText(secret)
		if err != nil {
			h.renderError(w, http.StatusInternalServerError, "Internal Error", "An unexpected error occurred.")
			return
		}
		h.store.Delete(id)
		h.templates.ExecuteTemplate(w, "show.html", text)
	default:
		w.Header().Set("Allow", "GET, POST")
		h.renderError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Method not allowed.")
	}
}

func (h *Handler) renderReveal(w http.ResponseWriter, id string) {
	token, err := h.generateCSRFToken()
	if err != nil {
		http.Error(w, "Could not generate CSRF token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})

	h.templates.ExecuteTemplate(w, "reveal.html", struct {
		ID        string
		CSRFToken string
	}{ID: id, CSRFToken: token})
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
