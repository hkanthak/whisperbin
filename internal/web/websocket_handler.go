package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // für Dev; in Prod auf Origin prüfen!
	},
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
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

	conn.WriteMessage(websocket.TextMessage, []byte(text))
	h.store.Delete(id)
}
