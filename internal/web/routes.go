package web

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"whisperbin/internal/storage"
)

type Handler struct {
	store         *storage.Store
	templates     *template.Template
	allowedOrigin string
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

	return &Handler{
		store:         store,
		templates:     tmpl,
		allowedOrigin: allowedOrigin,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.formHandler)
	mux.HandleFunc("/secret", h.createHandler)
	mux.HandleFunc("/confirm/", h.confirmHandler)
	mux.HandleFunc("/status/", h.statusHandler)
	mux.HandleFunc("/ws", h.WebSocketHandler)
	return mux
}
