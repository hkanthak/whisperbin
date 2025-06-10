package web

import (
	"html"
	"net/http"

	"github.com/gorilla/websocket"
)

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == h.allowedOrigin
		},
	}

	id := r.URL.Query().Get("id")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sec, err := h.store.WaitForUnlock(id)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error: "+err.Error()))
		return
	}

	text, err := h.store.DecryptSecretText(sec)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("error: decryption failed"))
		return
	}

	text = html.EscapeString(text)

	conn.WriteMessage(websocket.TextMessage, []byte(text))
	h.store.Delete(id)
}
