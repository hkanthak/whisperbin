package web

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"whisperbin/internal/storage"
)

func TestWebSocket_ReceivesSecretOnce(t *testing.T) {
	store := storage.NewStore()
	id, code, err := store.Save("via websocket", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	u := url.URL{
		Scheme: "ws",
		Host:   server.Listener.Addr().String(),
		Path:   "/ws",
	}
	q := u.Query()
	q.Set("id", id)
	u.RawQuery = q.Encode()

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("WebSocket connect failed: %v", err)
	}
	defer ws.Close()

	go func() {
		time.Sleep(50 * time.Millisecond)
		store.Confirm(id, code)
	}()

	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("WebSocket read failed: %v", err)
	}

	if string(msg) != "via websocket" {
		t.Errorf("Expected message %q, got %q", "via websocket", msg)
	}

	_, err = store.Get(id)
	if err == nil {
		t.Error("Expected secret to be deleted after WebSocket read")
	}
}
