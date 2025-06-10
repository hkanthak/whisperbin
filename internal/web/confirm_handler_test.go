package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	form.Add("code", "wrong")
	form.Add("csrf_token", csrfToken)

	postReq, _ := http.NewRequest("POST", server.URL+"/confirm/"+id, strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	postResp, err := client.Do(postReq)
	if err != nil {
		t.Fatal(err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", postResp.StatusCode)
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
	form.Add("code", code)
	form.Add("csrf_token", csrfToken)

	postReq, _ := http.NewRequest("POST", server.URL+"/confirm/"+id, strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})

	postResp, err := client.Do(postReq)
	if err != nil {
		t.Fatal(err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", postResp.StatusCode)
	}
}
