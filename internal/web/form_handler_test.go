package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"whisperbin/internal/storage"
)

func TestCreateHandler_NonSecure(t *testing.T) {
	store := storage.NewStore()
	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)

	server := httptest.NewServer(h.Routes())
	defer server.Close()

	client := &http.Client{}
	getReq, _ := http.NewRequest("GET", server.URL+"/", nil)
	getResp, err := client.Do(getReq)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()

	var csrfToken string
	for _, cookie := range getResp.Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
			break
		}
	}

	if csrfToken == "" {
		t.Fatal("CSRF token not found")
	}

	form := url.Values{}
	form.Add("secret", "my test secret")
	form.Add("ttl", "10")
	form.Add("csrf_token", csrfToken)

	postReq, _ := http.NewRequest("POST", server.URL+"/secret", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	postResp, err := client.Do(postReq)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", postResp.StatusCode)
	}
}

func TestCreateHandler_Secure(t *testing.T) {
	store := storage.NewStore()
	tmpl := projectRootPath("ui/templates/*.html")
	h := NewHandlerWithTemplates(store, tmpl)

	server := httptest.NewServer(h.Routes())
	defer server.Close()

	client := &http.Client{}
	getReq, _ := http.NewRequest("GET", server.URL+"/", nil)
	getResp, err := client.Do(getReq)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()

	var csrfToken string
	for _, cookie := range getResp.Cookies() {
		if cookie.Name == "csrf_token" {
			csrfToken = cookie.Value
			break
		}
	}

	if csrfToken == "" {
		t.Fatal("CSRF token not found")
	}

	form := url.Values{}
	form.Add("secret", "secure secret")
	form.Add("ttl", "5")
	form.Add("secure", "on")
	form.Add("csrf_token", csrfToken)

	postReq, _ := http.NewRequest("POST", server.URL+"/secret", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	postResp, err := client.Do(postReq)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", postResp.StatusCode)
	}
}
