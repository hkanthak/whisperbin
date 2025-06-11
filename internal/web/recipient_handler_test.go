package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"whisperbin/internal/storage"
)

func TestGetHandler_ShowsSecretAndDeletes(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("visible once", 5, false)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, "visible once") {
		t.Error("Expected secret to be shown")
	}

	resp2, _ := http.Get(server.URL + "/" + id)
	if resp2.StatusCode != http.StatusNotFound {
		t.Error("Expected 404 on second view")
	}
}

func TestGetHandler_SecureModeWait(t *testing.T) {
	store := storage.NewStore()
	id, code, err := store.Save("locked secret", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if !strings.Contains(body, code) {
		t.Errorf("Expected code %q in waiting page", code)
	}
	if strings.Contains(body, "locked secret") {
		t.Error("Secret should not be shown before unlock")
	}
}

func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(body)
}
