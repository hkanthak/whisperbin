package web

import (
	"html/template"
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
	tmpl := template.Must(template.ParseGlob(pattern))

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
