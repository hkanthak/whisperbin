package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"whisperbin/internal/storage"
)

func TestStatusHandler_WaitingExpectedTrueEvenBeforeClientConnect(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("locked", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/status/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"waiting": true`) {
		t.Errorf("Expected waiting=true, got: %s", body)
	}
}

func TestStatusHandler_WaitingTrue(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("locked", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	go store.WaitForUnlock(id)

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/status/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"waiting": true`) {
		t.Errorf("Expected waiting=true, got: %s", body)
	}
}

func TestStatusHandler_NotFound(t *testing.T) {
	store := storage.NewStore()

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, _ := http.Get(server.URL + "/status/does-not-exist")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.StatusCode)
	}
}
