package web

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"whisperbin/internal/storage"
)

func TestSSE_ReceivesSecretOnce(t *testing.T) {
	store := storage.NewStore()
	id, code, err := store.Save("via sse", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	go func() {
		time.Sleep(50 * time.Millisecond)
		store.Confirm(id, code, "127.0.0.1")
	}()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/sse?id="+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", h.allowedOrigin)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE connect failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var received string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			received = strings.TrimPrefix(line, "data: ")
			break
		}
	}

	if received != "via sse" {
		t.Errorf("Expected %q, got %q", "via sse", received)
	}

	_, err = store.Get(id)
	if err == nil {
		t.Error("Expected secret to be deleted after SSE delivery")
	}
}

func TestSSE_RejectsForbiddenOrigin(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("should not leak", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/sse?id="+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://evil.example.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}
