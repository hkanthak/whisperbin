package web

import (
	"fmt"
	"html"
	"net/http"
	"strings"
)

func (h *Handler) SSEHandler(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" && origin != h.allowedOrigin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sec, err := h.store.WaitForUnlock(r.Context(), id)
	if err != nil {
		fmt.Fprintf(w, "data: error: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	text, err := h.store.DecryptSecretText(sec)
	if err != nil {
		fmt.Fprintf(w, "data: error: decryption failed\n\n")
		flusher.Flush()
		return
	}

	text = strings.ReplaceAll(text, "\n", "\\n")
	text = html.EscapeString(text)

	fmt.Fprintf(w, "data: %s\n\n", text)
	flusher.Flush()

	h.store.Delete(id)
}
