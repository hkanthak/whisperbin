package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"whisperbin/internal/storage"
)

func TestGetHandler_ShowsRevealPageWithoutConsuming(t *testing.T) {
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
	if strings.Contains(body, "visible once") {
		t.Error("Secret should not be shown on GET")
	}
	if !strings.Contains(body, "Reveal Secret") {
		t.Error("Expected reveal form on GET")
	}

	resp2, err := http.Get(server.URL + "/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Error("Expected secret to survive repeated GETs")
	}
}

func TestPostHandler_RevealsSecretAndDeletes(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("visible once", 5, false)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	resp, body := revealPost(t, server.URL, id, getRevealToken(t, server.URL, id))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "visible once") {
		t.Error("Expected secret to be shown after reveal")
	}

	resp2, _ := http.Get(server.URL + "/" + id)
	if resp2.StatusCode != http.StatusNotFound {
		t.Error("Expected 404 after reveal")
	}
}

func TestPostHandler_InvalidCSRFDoesNotConsume(t *testing.T) {
	store := storage.NewStore()
	id, _, err := store.Save("still here", 5, false)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)
	server := httptest.NewServer(h.Routes())
	defer server.Close()

	form := url.Values{}
	form.Add("csrf_token", "wrong")
	resp, err := http.Post(server.URL+"/"+id, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected 403, got %d", resp.StatusCode)
	}

	resp2, body := revealPost(t, server.URL, id, getRevealToken(t, server.URL, id))
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK || !strings.Contains(body, "still here") {
		t.Error("Expected secret to survive a rejected POST")
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

func getRevealToken(t *testing.T, baseURL, id string) string {
	t.Helper()
	resp, err := http.Get(baseURL + "/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" {
			return cookie.Value
		}
	}
	t.Fatal("CSRF token not found")
	return ""
}

func revealPost(t *testing.T, baseURL, id, token string) (*http.Response, string) {
	t.Helper()
	form := url.Values{}
	form.Add("csrf_token", token)

	req, _ := http.NewRequest("POST", baseURL+"/"+id, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp, readBody(t, resp)
}

func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(body)
}
