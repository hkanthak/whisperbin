package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"whisperbin/internal/storage"
)

func TestCreateHandler_NonSecure(t *testing.T) {
	store := storage.NewStore()
	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)

	server := httptest.NewServer(h.Routes())
	defer server.Close()

	form := url.Values{}
	form.Add("secret", "my test secret")
	form.Add("ttl", "10")

	resp, err := http.PostForm(server.URL+"/secret", form)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestCreateHandler_Secure(t *testing.T) {
	store := storage.NewStore()
	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)

	server := httptest.NewServer(h.Routes())
	defer server.Close()

	form := url.Values{}
	form.Add("secret", "secure secret")
	form.Add("ttl", "5")
	form.Add("secure", "on")

	resp, err := http.PostForm(server.URL+"/secret", form)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
