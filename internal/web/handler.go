package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"whisperbin/internal"
	"whisperbin/internal/storage"
)

type Handler struct {
	store         *storage.Store
	templates     *template.Template
	allowedOrigin string
	ipLimiter     *ipLimiter
}

func NewHandler(store *storage.Store) *Handler {
	root, err := os.Getwd()
	if err != nil {
		panic("could not resolve working dir")
	}
	path := filepath.Join(root, "ui", "templates", "*.html")
	return NewHandlerWithTemplates(store, path)
}

func NewHandlerWithTemplates(store *storage.Store, pattern string) *Handler {
	tmpl := template.New("").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, _ := values[i].(string)
				dict[key] = values[i+1]
			}
			return dict
		},
	})
	tmpl = template.Must(tmpl.ParseGlob(pattern))

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:8080"
	}

	ipLimiter := newIPLimiter(internal.RateLimiterRate, internal.RateLimiterBurst)

	return &Handler{
		store:         store,
		templates:     tmpl,
		allowedOrigin: allowedOrigin,
		ipLimiter:     ipLimiter,
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

func (h *Handler) renderError(w http.ResponseWriter, status int, title string, message string) {
	w.WriteHeader(status)
	h.templates.ExecuteTemplate(w, "error.html", struct {
		Title   string
		Message string
	}{
		Title:   title,
		Message: message,
	})
}

func (h *Handler) renderSuccess(w http.ResponseWriter, title string, message string) {
	h.templates.ExecuteTemplate(w, "success.html", struct {
		Title   string
		Message string
	}{
		Title:   title,
		Message: message,
	})
}
