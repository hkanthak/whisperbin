package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"whisperbin/internal/storage"
)

func TestConfirmHandler_InvalidCode(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("locked", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	form := url.Values{}
	form.Add("code", "wrong")
	resp, err := http.PostForm(server.URL+"/confirm/"+id, form)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", resp.StatusCode)
	}
}

func TestConfirmHandler_ValidCode(t *testing.T) {
	store := storage.NewStore()
	id, code, err := store.Save("to be unlocked", 5, true)
	if err != nil {
		t.Fatal(err)
	}

	go store.WaitForUnlock(id)

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	form := url.Values{}
	form.Add("code", code)
	resp, err := http.PostForm(server.URL+"/confirm/"+id, form)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}
