package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.handleForm(w, r)
		} else {
			h.handleReveal(w, r)
		}
	})

	mux.HandleFunc("/secret", h.handleCreate)

	return mux
}

func (h *Handler) handleForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := strings.TrimSpace(r.FormValue("secret"))
	ttl := r.FormValue("ttl")

	if text == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	minutes := 10
	if ttl != "" {
		fmt.Sscanf(ttl, "%d", &minutes)
	}

	id, err := h.store.Save(text, minutes)
	if err != nil {
		http.Error(w, "Could not save secret", http.StatusInternalServerError)
		return
	}

	link := "/" + id
	h.templates.ExecuteTemplate(w, "created.html", link)
}

func (h *Handler) handleReveal(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	secret, err := h.store.LoadAndDelete(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.templates.ExecuteTemplate(w, "show.html", secret)
}
