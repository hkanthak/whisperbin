package web

import (
	"net/http"
	"strings"
)

func (h *Handler) formHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.recipientHandler(w, r)
		return
	}
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

	if secure {
		h.templates.ExecuteTemplate(w, "created_secure.html", struct{ ID string }{ID: id})
	} else {
		h.templates.ExecuteTemplate(w, "created.html", "/"+id)
	}
}
