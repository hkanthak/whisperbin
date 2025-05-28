package web

import (
	"html/template"
	"net/http"

	"whisperbin/internal/storage"
)

type Handler struct {
	store     *storage.Store
	templates *template.Template
}

func NewHandler(store *storage.Store) *Handler {
	tmpl := template.Must(template.ParseGlob("internal/web/templates/*.html"))
	return &Handler{
		store:     store,
		templates: tmpl,
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
