package web

import (
	"crypto/rand"
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

	token := h.generateCSRFToken()
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

	token := h.generateCSRFToken()
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

func (h *Handler) generateCSRFToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("could not generate CSRF token")
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handler) validateCSRF(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return false
	}
	formToken := r.FormValue("csrf_token")
	return subtleConstantTimeCompare(cookie.Value, formToken)
}

func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
