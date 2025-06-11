package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) formHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.recipientHandler(w, r)
		return
	}

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

	h.templates.ExecuteTemplate(w, "index.html", struct{ CSRFToken string }{CSRFToken: token})
}

func (h *Handler) createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !h.validateCSRF(w, r) {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	text := strings.TrimSpace(r.FormValue("secret"))
	if len(text) > 10240 {
		http.Error(w, "Secret too large", http.StatusRequestEntityTooLarge)
		return
	}

	ttl := 10
	if ttl < 1 {
		ttl = 1
	} else if ttl > 1440 {
		ttl = 1440
	}
	secure := r.FormValue("secure") == "on"

	id, _, err := h.store.Save(text, ttl, secure)
	if err != nil {
		http.Error(w, "Could not save secret", http.StatusInternalServerError)
		return
	}

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

	if secure {
		h.templates.ExecuteTemplate(w, "created_secure.html", struct {
			ID        string
			CSRFToken string
		}{ID: id, CSRFToken: token})
	} else {
		h.templates.ExecuteTemplate(w, "created.html", "/"+id)
	}
}

func (h *Handler) generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *Handler) validateCSRF(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return false
	}
	formToken := r.FormValue("csrf_token")
	if len(cookie.Value) != len(formToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(formToken)) == 1
}
